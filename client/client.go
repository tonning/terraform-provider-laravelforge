package client

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// HostURL - Default Laravel Forge API URL
const HostURL string = "https://forge.laravel.com/api/v1"

type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

func NewClient(host, token *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    HostURL,
		Token:      *token,
	}

	if host != nil {
		c.HostURL = *host
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[INFO] [LARAVELFORGE] [doRequest] Request %#v", req)
	log.Printf("[INFO] [LARAVELFORGE] [doRequest] URL %s", c.HostURL)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	//log.Printf("[INFO] [LARAVELFORGE] [doRequest] Body: %v", string(body))
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}

func (c *Client) doRequestEmptyBody(req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	log.Printf("[INFO] [LARAVELFORGE] [doRequest] Request %#v", req)
	log.Printf("[INFO] [LARAVELFORGE] [doRequest] URL %s", c.HostURL)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status: %d", res.StatusCode)
	}

	return err
}
