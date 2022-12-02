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

func (c *Client) GetServer(serverId string) (*Server, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetServer] ServerId: %s", serverId)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] [LARAVELFORGE:GetServer] Server: %#v, Body: %#v", &server, body)

	return &server.Server, nil
}

func (c *Client) CreateServer(createServer *ServerCreateRequest) (*ServerResponse, error) {
	log.Printf("[INFO] [LARAVELFORGE:CreateServer]")
	rb, err := json.Marshal(createServer)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/servers", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] [LARAVELFORGE:CreateServer] Body: %#v", string(body))
	log.Printf("[INFO] [LARAVELFORGE:CreateServer] Server: %#v", server)

	return &server, nil
}

func (c *Client) UpdateServer(serverId string, serverUpdates ServerUpdateRequest) (*Server, diag.Diagnostics) {
	rb, err := json.Marshal(serverUpdates)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}
	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	server := ServerResponse{}
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &server.Server, nil
}

func (c *Client) DeleteServer(serverId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s", c.HostURL, serverId), nil)
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
