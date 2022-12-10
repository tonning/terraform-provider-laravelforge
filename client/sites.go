package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"net/http"
	"strings"
)

// GetSite - Returns a specific site in a project
func (c *Client) GetSite(serverId string, siteId string) (*Site, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/sites/%s", c.HostURL, serverId, siteId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetSite] SiteId: %s", siteId)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	site := SiteGet{}
	err = json.Unmarshal(body, &site)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] [LARAVELFORGE:GetSite] Site: %#v, Body: %#v", &site, body)

	return &site.Site, nil
}

func (c *Client) CreateSite(serverId string, createSite *SiteCreateRequest) (*Site, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateSite]")
	rb, err := json.Marshal(createSite)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/sites", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	site := SiteGet{}
	err = json.Unmarshal(body, &site)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &site.Site, nil
}

func (c *Client) UpdateSite(serverId string, siteId string, siteUpdates SiteUpdateRequest) (*Site, diag.Diagnostics) {
	rb, err := json.Marshal(siteUpdates)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/servers/%s/sites/%s", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	site := SiteGet{}
	err = json.Unmarshal(body, &site)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &site.Site, nil
}

func (c *Client) UpdateSitePhpVersion(serverId string, siteId string, phpVersion SiteUpdatePhpVersion) (*Site, diag.Diagnostics) {
	rb, err := json.Marshal(phpVersion)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/servers/%s/sites/%s/php", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	err = c.doRequestEmptyBody(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	site, err := c.GetSite(serverId, siteId)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return site, nil
}

func (c *Client) DeleteSite(serverId string, siteId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/sites/%s", c.HostURL, serverId, siteId), nil)
	if err != nil {
		return err
	}
	body, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
