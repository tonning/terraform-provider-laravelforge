package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HostURL - Default Laravel Forge API URL
const HostURL string = "https://forge.laravel.com/api"

const maxRateLimitRetries = 10

// DefaultMinRequestInterval keeps the client below Forge's limit of 60 requests per minute.
const DefaultMinRequestInterval = 1200 * time.Millisecond

type Client struct {
	HostURL      string
	HTTPClient   *http.Client
	Token        string
	Organization string

	// MinRequestInterval is the minimum delay between two requests, shared by
	// all concurrent callers. Zero disables throttling.
	MinRequestInterval time.Duration

	mu       sync.Mutex
	nextSlot time.Time
}

// throttle blocks until the caller may send a request.
func (c *Client) throttle() {
	if c.MinRequestInterval <= 0 {
		return
	}

	c.mu.Lock()
	now := time.Now()
	slot := c.nextSlot
	if slot.Before(now) {
		slot = now
	}
	c.nextSlot = slot.Add(c.MinRequestInterval)
	c.mu.Unlock()

	if d := time.Until(slot); d > 0 {
		time.Sleep(d)
	}
}

// pauseUntil delays all callers, used after a rate limit response.
func (c *Client) pauseUntil(t time.Time) {
	c.mu.Lock()
	if t.After(c.nextSlot) {
		c.nextSlot = t
	}
	c.mu.Unlock()
}

// retryDelay derives how long to wait after a 429 response.
func retryDelay(res *http.Response, attempt int) time.Duration {
	if s, err := strconv.Atoi(res.Header.Get("Retry-After")); err == nil && s >= 0 {
		return time.Duration(s)*time.Second + time.Second
	}
	if s, err := strconv.ParseInt(res.Header.Get("X-RateLimit-Reset"), 10, 64); err == nil {
		if d := time.Until(time.Unix(s, 0)); d > 0 && d < 5*time.Minute {
			return d + time.Second
		}
	}

	d := 15 * time.Second << uint(attempt)
	if d > 2*time.Minute {
		d = 2 * time.Minute
	}

	return d
}

// APIError is returned for every non 2xx response.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
}

// IsNotFound reports whether err is a 404 response from the Forge API.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// NewClient creates a client for the organization scoped Forge API. When no
// organization is given and the token only has access to a single one, that
// organization is used.
func NewClient(host, token, organization *string) (*Client, error) {
	c := Client{
		HTTPClient:         &http.Client{Timeout: 30 * time.Second},
		HostURL:            HostURL,
		Token:              *token,
		MinRequestInterval: DefaultMinRequestInterval,
	}

	if host != nil && *host != "" {
		c.HostURL = strings.TrimRight(*host, "/")
	}

	if organization != nil {
		c.Organization = strings.TrimSpace(*organization)
	}

	if c.Organization == "" {
		slug, err := c.discoverOrganization()
		if err != nil {
			return nil, err
		}
		c.Organization = slug
	}

	return &c, nil
}

func (c *Client) discoverOrganization() (string, error) {
	items, err := c.list(fmt.Sprintf("%s/orgs", c.HostURL), nil)
	if err != nil {
		return "", fmt.Errorf("unable to list organizations: %w", err)
	}

	slugs := make([]string, 0, len(items))
	for _, item := range items {
		var attrs struct {
			Slug string `json:"slug"`
		}
		if err := json.Unmarshal(item.Attributes, &attrs); err != nil {
			return "", err
		}
		slugs = append(slugs, attrs.Slug)
	}

	switch len(slugs) {
	case 0:
		return "", errors.New("the Laravel Forge token has no access to any organization")
	case 1:
		return slugs[0], nil
	default:
		return "", fmt.Errorf("the Laravel Forge token has access to multiple organizations (%s); set the provider `organization` argument or LARAVELFORGE_ORGANIZATION", strings.Join(slugs, ", "))
	}
}

// orgURL builds a URL below /orgs/{organization}.
func (c *Client) orgURL(format string, args ...interface{}) string {
	return fmt.Sprintf("%s/orgs/%s", c.HostURL, url.PathEscape(c.Organization)) + fmt.Sprintf(format, args...)
}

// do executes a request and returns the body of any 2xx response.
func (c *Client) do(method, endpoint string, payload interface{}) ([]byte, error) {
	var rb []byte
	if payload != nil {
		var err error
		rb, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	for attempt := 0; ; attempt++ {
		var reader io.Reader
		if rb != nil {
			reader = bytes.NewReader(rb)
		}

		req, err := http.NewRequest(method, endpoint, reader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		c.throttle()
		res, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			return nil, err
		}

		if res.StatusCode == http.StatusTooManyRequests && attempt < maxRateLimitRetries {
			wait := retryDelay(res, attempt)
			log.Printf("[INFO] [LARAVELFORGE] Rate limited, retrying in %s", wait)
			c.pauseUntil(time.Now().Add(wait))
			continue
		}

		log.Printf("[DEBUG] [LARAVELFORGE] %s %s -> %d", method, endpoint, res.StatusCode)

		if res.StatusCode < 200 || res.StatusCode > 299 {
			return nil, &APIError{StatusCode: res.StatusCode, Body: string(body)}
		}

		return body, nil
	}
}

// resource is a JSON:API resource object.
type resource struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Attributes json.RawMessage `json:"attributes"`
}

type singleEnvelope struct {
	Data resource        `json:"data"`
	Meta json.RawMessage `json:"meta"`
}

type listEnvelope struct {
	Data []resource `json:"data"`
	Meta struct {
		NextCursor *string `json:"next_cursor"`
	} `json:"meta"`
}

// decodeSingle decodes a single JSON:API resource into attrs and returns its
// id and the optional top level meta object. An empty body yields an empty id.
func decodeSingle(body []byte, attrs interface{}) (string, json.RawMessage, error) {
	if len(bytes.TrimSpace(body)) == 0 {
		return "", nil, nil
	}

	var env singleEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		return "", nil, err
	}

	if attrs != nil && len(env.Data.Attributes) > 0 {
		if err := json.Unmarshal(env.Data.Attributes, attrs); err != nil {
			return "", nil, err
		}
	}

	return env.Data.ID, env.Meta, nil
}

func unmarshalAttributes(r resource, attrs interface{}) error {
	if len(r.Attributes) == 0 {
		return nil
	}

	return json.Unmarshal(r.Attributes, attrs)
}

// list returns every item of a paginated collection.
func (c *Client) list(endpoint string, query url.Values) ([]resource, error) {
	var items []resource
	cursor := ""

	for {
		q := url.Values{}
		for k, v := range query {
			q[k] = v
		}
		q.Set("page[size]", "100")
		if cursor != "" {
			q.Set("page[cursor]", cursor)
		}

		body, err := c.do(http.MethodGet, endpoint+"?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}

		var env listEnvelope
		if err := json.Unmarshal(body, &env); err != nil {
			return nil, err
		}
		items = append(items, env.Data...)

		if env.Meta.NextCursor == nil || *env.Meta.NextCursor == "" {
			return items, nil
		}
		cursor = *env.Meta.NextCursor
	}
}

// PollInterval is the default delay between status checks.
var PollInterval = 10 * time.Second

// WaitUntil calls fn until it reports done, returns an error, or the timeout elapses.
func WaitUntil(timeout time.Duration, fn func() (bool, error)) error {
	deadline := time.Now().Add(timeout)

	for {
		done, err := fn()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		if time.Now().Add(PollInterval).After(deadline) {
			return errors.New("timed out waiting for Laravel Forge")
		}
		time.Sleep(PollInterval)
	}
}

// FlexString decodes JSON strings and numbers (and null) into a string.
type FlexString string

func (f *FlexString) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" || s == "" {
		*f = ""
		return nil
	}
	if strings.HasPrefix(s, "\"") {
		var v string
		if err := json.Unmarshal(b, &v); err != nil {
			return err
		}
		*f = FlexString(v)
		return nil
	}
	*f = FlexString(s)
	return nil
}
