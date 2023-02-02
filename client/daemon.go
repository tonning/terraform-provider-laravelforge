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

func (c *Client) GetDaemon(serverId string, daemonId string) (*Daemon, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/daemons/%s", c.HostURL, serverId, daemonId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetDaemon] DaemonId: %s", daemonId)
	if err != nil {
		return nil, err
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	daemon := DaemonResponse{}
	err = json.Unmarshal(body, &daemon)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] [LARAVELFORGE:GetDaemon] Daemon: %#v, Body: %#v", &daemon, body)

	return &daemon.Daemon, nil
}

func (c *Client) CreateDaemon(serverId string, createDaemon *CreateDaemonRequest) (*Daemon, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateDaemon]")
	rb, err := json.Marshal(createDaemon)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/daemons", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	daemon := DaemonResponse{}
	err = json.Unmarshal(body, &daemon)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &daemon.Daemon, nil
}

func (c *Client) DeleteDaemon(serverId string, daemonId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/daemons/%s", c.HostURL, serverId, daemonId), nil)
	if err != nil {
		return err
	}
	body, err, _ := c.doRequest(req)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
