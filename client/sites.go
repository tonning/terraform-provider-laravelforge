package client

import (
	"net/http"
)

type Site struct {
	Id            string `json:"-"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	User          string `json:"user"`
	WebDirectory  string `json:"web_directory"`
	RootDirectory string `json:"root_directory"`
	Wildcards     *bool  `json:"wildcards"`
}

type SiteCreateRequest struct {
	Domain       string
	ProjectType  string
	WebDirectory string
	Username     string
	PhpVersion   string
	Wildcards    bool
}

type SiteUpdateRequest struct {
	Directory  string `json:"directory,omitempty"`
	PhpVersion string `json:"php_version,omitempty"`
}

type Domain struct {
	Id                      string `json:"-"`
	Name                    string `json:"name"`
	Type                    string `json:"type"`
	Status                  string `json:"status"`
	WwwRedirectType         string `json:"www_redirect_type"`
	AllowWildcardSubdomains bool   `json:"allow_wildcard_subdomains"`
}

func (r *SiteCreateRequest) payload() map[string]interface{} {
	p := map[string]interface{}{
		"type":                      r.ProjectType,
		"domain_mode":               "custom",
		"name":                      r.Domain,
		"www_redirect_type":         "none",
		"allow_wildcard_subdomains": r.Wildcards,
	}

	if r.WebDirectory != "" {
		p["web_directory"] = r.WebDirectory
	}
	if r.PhpVersion != "" {
		p["php_version"] = r.PhpVersion
	}
	if r.Username != "" && r.Username != "forge" {
		p["is_isolated"] = true
		p["isolated_user"] = r.Username
	}

	return p
}

func decodeSite(body []byte) (*Site, error) {
	site := Site{}
	id, _, err := decodeSingle(body, &site)
	if err != nil {
		return nil, err
	}
	site.Id = id

	return &site, nil
}

// GetSite - Returns a specific site
func (c *Client) GetSite(siteId string) (*Site, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/sites/%s", siteId), nil)
	if err != nil {
		return nil, err
	}

	return decodeSite(body)
}

func (c *Client) CreateSite(serverId string, createSite *SiteCreateRequest) (*Site, error) {
	body, err := c.do(http.MethodPost, c.orgURL("/servers/%s/sites", serverId), createSite.payload())
	if err != nil {
		return nil, err
	}

	return decodeSite(body)
}

func (c *Client) UpdateSite(serverId string, siteId string, siteUpdates SiteUpdateRequest) error {
	_, err := c.do(http.MethodPut, c.orgURL("/servers/%s/sites/%s", serverId, siteId), siteUpdates)

	return err
}

func (c *Client) DeleteSite(serverId string, siteId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/sites/%s", serverId, siteId), nil)

	return err
}

func (c *Client) ListDomains(serverId string, siteId string) ([]Domain, error) {
	items, err := c.list(c.orgURL("/servers/%s/sites/%s/domains", serverId, siteId), nil)
	if err != nil {
		return nil, err
	}

	domains := make([]Domain, 0, len(items))
	for _, item := range items {
		domain := Domain{}
		if err := unmarshalAttributes(item, &domain); err != nil {
			return nil, err
		}
		domain.Id = item.ID
		domains = append(domains, domain)
	}

	return domains, nil
}

func (c *Client) CreateDomain(serverId string, siteId string, name string, wildcards bool) error {
	payload := map[string]interface{}{
		"name":                      name,
		"allow_wildcard_subdomains": wildcards,
		"www_redirect_type":         "none",
	}
	_, err := c.do(http.MethodPost, c.orgURL("/servers/%s/sites/%s/domains", serverId, siteId), payload)

	return err
}

func (c *Client) UpdateDomain(serverId string, siteId string, domainId string, wwwRedirectType string, wildcards bool) error {
	if wwwRedirectType == "" {
		wwwRedirectType = "none"
	}

	payload := map[string]interface{}{
		"www_redirect_type":         wwwRedirectType,
		"allow_wildcard_subdomains": wildcards,
	}
	_, err := c.do(http.MethodPatch, c.orgURL("/servers/%s/sites/%s/domains/%s", serverId, siteId, domainId), payload)

	return err
}

func (c *Client) DeleteDomain(serverId string, siteId string, domainId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/sites/%s/domains/%s", serverId, siteId, domainId), nil)

	return err
}
