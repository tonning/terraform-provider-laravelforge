package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (c *Client) GetKey(serverId string, keyId string) (*Key, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/keys/%s", c.HostURL, serverId, keyId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetKey] KeyID: %s", keyId)
	if err != nil {
		return nil, err
	}

	body, err, _ := c.doRequest(req)
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

func (c *Client) CreateKey(serverId string, keyCreateRequest *KeyCreateRequest, retry bool) (*Key, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateKey]")
	rb, err := json.Marshal(keyCreateRequest)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/keys", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)

	if err != nil && err.Error() == "status: 422, body: {\"name\":[\"The name has already been taken.\"]}" && keyCreateRequest.Overwrite == true && retry == true {
		log.Printf("[DEBUG] [CreateKey] Key already exists.]")
		key, searchedKeyErr := c.SearchKeyByName(serverId, keyCreateRequest.Name)
		log.Printf("[DEBUG] Searched key: %#v, Server ID: %s", key, serverId)

		if searchedKeyErr != nil {
			log.Printf("[DEBUG] [CreateKey] error thrown. searchedKeyErr != nil")
			return nil, diag.Errorf("Whoops: %s", searchedKeyErr)
		}

		if key != nil {
			log.Printf("[DEBUG] [CreateKey] about to delete existing key")
			err := c.DeleteKey(serverId, strconv.Itoa(key.Id))

			if err != nil {
				log.Printf("[ERROR] [CreateKey] Error deleting key: %s", err)
				return nil, diag.Errorf("Whoops: %s", err)
			}
		}

		log.Printf("[DEBUG] [CreateKey] wait 10 seconds")
		time.Sleep(time.Second * 10)
		log.Printf("[DEBUG] [CreateKey] waited 10 seconds")

		log.Printf("[DEBUG] [CreateKey] about to create new key")

		return c.CreateKey(serverId, keyCreateRequest, false)
	}

	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
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
	log.Printf("[DEBUG] [DeleteKey] KeyId: %s, Server ID: %s", keyId, serverId)
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/keys/%s", c.HostURL, serverId, keyId), nil)
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

func (c *Client) ListKeys(serverId string) ([]Key, diag.Diagnostics) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/keys", c.HostURL, serverId), nil)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	log.Printf("[DEBUG] [List keys] - body: %#v, Server ID: %s", string(body), serverId)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	// Parse the JSON body into the KeysResponse struct
	var keysResponse KeysResponse
	err = json.Unmarshal(body, &keysResponse)

	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	log.Printf("[DEBUG] [List keys] keys: %#v, Server ID: %s", keysResponse, serverId)
	return keysResponse.Keys, nil
}

func (c *Client) SearchKeyByName(serverId string, keyName string) (*Key, diag.Diagnostics) {
	keys, err := c.ListKeys(serverId)

	log.Printf("[DEBUG] [SearchKeyByName] keys: %#v, Key Name: %s, Server ID: %s", keys, keyName, serverId)

	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	for _, key := range keys {
		log.Printf("[DEBUG] [SearchKeyByName] iteration key: %#v, Key Name: %s", key, keyName)
		if key.Name == keyName {
			return &key, nil
		}
	}

	return nil, nil
}
