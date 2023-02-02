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

func (c *Client) GetServer(serverId string) (*Server, error, *http.Response) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetServer] ServerId: %s", serverId)
	if err != nil {
		return nil, err, nil
	}

	body, err, res := c.doRequest(req)
	if err != nil {
		if res.StatusCode == 404 {
			log.Printf("[INFO] [LARAVELFORGE:GetServer] Response: %#v", res)
		}
		return nil, err, res
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err, res
	}
	log.Printf("[INFO] [LARAVELFORGE:GetServer] Server: %#v, Body: %#v", &server, body)

	return &server.Server, nil, res
}

func (c *Client) CreateServer(createServer *ServerCreateRequest) (*ServerResponse, error, *http.Response) {
	log.Printf("[INFO] [LARAVELFORGE:CreateServer]")
	rb, err := json.Marshal(createServer)
	if err != nil {
		return nil, err, nil
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/servers", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err, nil
	}

	body, err, res := c.doRequest(req)
	log.Printf("[INFO] [LARAVELFORGE:CreateServer] Body: %#v, Error: %#v, Response: %#v", body, err, res)

	if err != nil {
		return nil, err, res
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err, res
	}

	log.Printf("[INFO] [LARAVELFORGE:CreateServer] Body: %#v", string(body))
	log.Printf("[INFO] [LARAVELFORGE:CreateServer] Server: %#v", server)

	return &server, nil, res
}

func (c *Client) UpdateServer(serverId string, serverUpdates ServerUpdateRequest) (*Server, diag.Diagnostics, *http.Response) {
	rb, err := json.Marshal(serverUpdates)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err), nil
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err), nil
	}

	body, err, res := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err), res
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err), res
	}

	return &server.Server, nil, res
}

func (c *Client) DeleteServer(serverId string) (error, *http.Response) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), nil)
	if err != nil {
		return err, nil
	}
	body, err, res := c.doRequest(req)
	if err != nil {
		return err, res
	}

	if string(body) != "" {
		return errors.New(string(body)), res
	}

	return nil, res
}

func (c *Client) EnableOpcache(serverId string) diag.Diagnostics {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/servers/%s/php/opcache", c.HostURL, serverId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.doRequestEmptyBody(req)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func (c *Client) DisableOpcache(serverId string) diag.Diagnostics {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/servers/%s/php/opcache", c.HostURL, serverId), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	err = c.doRequestEmptyBody(req)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
