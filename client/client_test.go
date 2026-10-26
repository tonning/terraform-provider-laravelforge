package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type recordedRequest struct {
	Method string
	Path   string
	Query  string
	Auth   string
	Body   map[string]interface{}
}

func newTestClient(t *testing.T, handler func(w http.ResponseWriter, r *http.Request, rec recordedRequest)) (*Client, *[]recordedRequest) {
	t.Helper()
	PollInterval = time.Millisecond

	var requests []recordedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := recordedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Auth: r.Header.Get("Authorization")}
		if raw, _ := io.ReadAll(r.Body); len(raw) > 0 {
			_ = json.Unmarshal(raw, &rec.Body)
		}
		requests = append(requests, rec)
		w.Header().Set("Content-Type", "application/vnd.api+json")
		handler(w, r, rec)
	}))
	t.Cleanup(srv.Close)

	token, org := "secret", "acme"
	c, err := NewClient(&srv.URL, &token, &org)
	if err != nil {
		t.Fatal(err)
	}

	return c, &requests
}

func TestNewClientDiscoversSingleOrganization(t *testing.T) {
	PollInterval = time.Millisecond
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orgs" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"data":[{"id":"org-1","type":"organizations","attributes":{"slug":"3z-brands"}}],"meta":{"next_cursor":null}}`)
	}))
	defer srv.Close()

	token := "secret"
	c, err := NewClient(&srv.URL, &token, nil)
	if err != nil {
		t.Fatal(err)
	}
	if c.Organization != "3z-brands" {
		t.Fatalf("organization = %q", c.Organization)
	}
}

func TestNewClientFailsWithMultipleOrganizations(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"data":[{"id":"1","attributes":{"slug":"a"}},{"id":"2","attributes":{"slug":"b"}}],"meta":{}}`)
	}))
	defer srv.Close()

	token := "secret"
	if _, err := NewClient(&srv.URL, &token, nil); err == nil || !strings.Contains(err.Error(), "multiple organizations") {
		t.Fatalf("expected multiple organizations error, got %v", err)
	}
}

func TestCreateCustomServerParsesProvisioning(t *testing.T) {
	c, requests := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"data":{"id":"42","type":"servers","attributes":{"name":"web-1","credential_id":null,"ssh_port":22,"is_ready":false,"ip_address":"1.2.3.4","provision_command":"curl | bash","sudo_password":"pw"}}}`)
	})

	server, err := c.CreateServer(&ServerCreateRequest{
		Name: "web-1", Provider: "custom", Type: "web", UbuntuVersion: "24.04", PhpVersion: "php83",
		IpAddress: "1.2.3.4", PrivateIpAddress: "10.0.0.4",
	})
	if err != nil {
		t.Fatal(err)
	}

	if server.Id != "42" || server.ProvisionCommand != "curl | bash" || server.SudoPassword != "pw" || server.IsReady {
		t.Fatalf("unexpected server %#v", server)
	}

	req := (*requests)[0]
	if req.Method != "POST" || req.Path != "/orgs/acme/servers" || req.Auth != "Bearer secret" {
		t.Fatalf("unexpected request %#v", req)
	}
	custom, _ := req.Body["custom"].(map[string]interface{})
	if custom["ip_address"] != "1.2.3.4" || custom["private_ip_address"] != "10.0.0.4" || req.Body["provider"] != "custom" {
		t.Fatalf("unexpected body %#v", req.Body)
	}
}

func TestCreateServerReadsProvisioningFromMeta(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"data":{"id":"7","attributes":{"credential_id":12}},"meta":{"provision_command":"cmd","sudo_password":"pw"}}`)
	})

	server, err := c.CreateServer(&ServerCreateRequest{Name: "x", Provider: "ocean2", Type: "web", Region: "ams3", Size: "s-1vcpu-1gb"})
	if err != nil {
		t.Fatal(err)
	}
	if server.ProvisionCommand != "cmd" || server.SudoPassword != "pw" || server.CredentialId != "12" {
		t.Fatalf("unexpected server %#v", server)
	}
}

func TestNotFoundIsDetectable(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := c.GetServer("1")
	if !IsNotFound(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestCreateKeyLooksUpAsyncKeyByNameAcrossPages(t *testing.T) {
	polls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		switch {
		case rec.Method == "POST":
			w.WriteHeader(http.StatusAccepted)
		case strings.Contains(rec.Query, "page%5Bcursor%5D=next"):
			polls++
			fmt.Fprint(w, `{"data":[{"id":"9","type":"keys","attributes":{"name":"laura","user":"forge","status":"installed"}}],"meta":{"next_cursor":null}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"1","type":"keys","attributes":{"name":"other"}}],"meta":{"next_cursor":"next"}}`)
		}
	})

	key, err := c.CreateKey("5", &KeyCreateRequest{Name: "laura", Key: "ssh-rsa AAA", Username: "forge"})
	if err != nil {
		t.Fatal(err)
	}
	if key.Id != "9" || key.User != "forge" || polls == 0 {
		t.Fatalf("unexpected key %#v", key)
	}
}

func TestCreateKeyOverwriteReplacesExistingKey(t *testing.T) {
	deleted, created := false, false
	c, requests := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		switch {
		case rec.Method == "POST" && !deleted:
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, `{"message":"The name has already been taken."}`)
		case rec.Method == "POST":
			created = true
			w.WriteHeader(http.StatusAccepted)
		case rec.Method == "DELETE":
			deleted = true
			w.WriteHeader(http.StatusAccepted)
		case created:
			fmt.Fprint(w, `{"data":[{"id":"4","attributes":{"name":"laura"}}],"meta":{}}`)
		case deleted:
			fmt.Fprint(w, `{"data":[],"meta":{}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"3","attributes":{"name":"laura"}}],"meta":{}}`)
		}
	})

	key, err := c.CreateKey("5", &KeyCreateRequest{Name: "laura", Key: "k", Overwrite: true})
	if err != nil {
		t.Fatal(err)
	}
	if key.Id != "4" {
		t.Fatalf("unexpected key %#v (%d requests)", key, len(*requests))
	}
}

func TestCreateSiteAndDomains(t *testing.T) {
	c, requests := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"data":{"id":"77","type":"sites","attributes":{"name":"example.com","status":"installing","user":"forge","web_directory":"/current/public","wildcards":true}}}`)
	})

	site, err := c.CreateSite("5", &SiteCreateRequest{Domain: "example.com", ProjectType: "php", WebDirectory: "/current/public", Username: "forge", PhpVersion: "php83", Wildcards: true})
	if err != nil {
		t.Fatal(err)
	}
	if site.Id != "77" || site.Wildcards == nil || !*site.Wildcards {
		t.Fatalf("unexpected site %#v", site)
	}

	body := (*requests)[0].Body
	if body["name"] != "example.com" || body["type"] != "php" || body["domain_mode"] != "custom" || body["web_directory"] != "/current/public" || body["allow_wildcard_subdomains"] != true {
		t.Fatalf("unexpected body %#v", body)
	}
	if _, isolated := body["is_isolated"]; isolated {
		t.Fatalf("forge user must not isolate the site: %#v", body)
	}
}

func TestRateLimitIsRetried(t *testing.T) {
	calls := 0
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, `{"data":{"id":"1","attributes":{"name":"s"}}}`)
	})

	server, err := c.GetServer("1")
	if err != nil || server.Name != "s" || calls != 2 {
		t.Fatalf("server=%v err=%v calls=%d", server, err, calls)
	}
}

func TestCertificateLookupAcrossDomains(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		switch rec.Path {
		case "/orgs/acme/servers/1/sites/2/domains":
			fmt.Fprint(w, `{"data":[{"id":"10","attributes":{"name":"a.com","type":"primary"}},{"id":"11","attributes":{"name":"b.com","type":"alias"}}],"meta":{}}`)
		case "/orgs/acme/servers/1/sites/2/domains/11/certificates/55":
			fmt.Fprint(w, `{"data":{"id":"55","attributes":{"type":"letsencrypt","status":"installed","active":true}}}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	certificate, err := c.FindCertificate("1", "2", "55")
	if err != nil {
		t.Fatal(err)
	}
	if certificate.DomainId != "11" || !certificate.Active {
		t.Fatalf("unexpected certificate %#v", certificate)
	}
}

func TestNumericAttributeIdsDoNotClashWithResourceIds(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request, rec recordedRequest) {
		switch {
		case strings.HasSuffix(rec.Path, "/servers/863132"):
			fmt.Fprint(w, `{"data":{"id":"863132","type":"servers","attributes":{"id":863132,"name":"web-ava","ssh_port":22,"credential_id":null}}}`)
		case strings.HasSuffix(rec.Path, "/sites/9"):
			fmt.Fprint(w, `{"data":{"id":"9","type":"sites","attributes":{"id":9,"name":"example.com","status":"installed"}}}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"3","attributes":{"id":3,"name":"k"}}],"meta":{}}`)
		}
	})

	server, err := c.GetServer("863132")
	if err != nil || server.Id != "863132" || server.Name != "web-ava" {
		t.Fatalf("server=%#v err=%v", server, err)
	}

	site, err := c.GetSite("9")
	if err != nil || site.Id != "9" {
		t.Fatalf("site=%#v err=%v", site, err)
	}

	keys, err := c.ListKeys("1")
	if err != nil || len(keys) != 1 || keys[0].Id != "3" {
		t.Fatalf("keys=%#v err=%v", keys, err)
	}
}
