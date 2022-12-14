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

func (c *Client) GetKey(serverId string, keyId string) (*Key, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/keys/%s", c.HostURL, serverId, keyId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetKey] KeyID: %s", keyId)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	key := KeyGet{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] [LARAVELFORGE:GetKey] Key: %#v, Body: %#v", &key, body)

	return &key.Key, nil
}

func (c *Client) CreateKey(serverId string, keyCreateRequest *KeyCreateRequest) (*Key, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateKey]")
	rb, err := json.Marshal(keyCreateRequest)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/keys", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	key := KeyGet{}
	err = json.Unmarshal(body, &key)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &key.Key, nil
}

func (c *Client) UpdateKey() error {
	return nil
}

func (c *Client) DeleteKey(serverId string, keyId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/keys/%s", c.HostURL, serverId, keyId), nil)
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
