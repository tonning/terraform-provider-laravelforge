package client

import (
	"net/http"
	"strings"
	"time"
)

type Key struct {
	Id     string
	Name   string `json:"name"`
	User   string `json:"user"`
	Status string `json:"status"`
}

type KeyCreateRequest struct {
	Name      string
	Key       string
	Username  string
	Overwrite bool
}

func (c *Client) GetKey(serverId string, keyId string) (*Key, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s/ssh-keys/%s", serverId, keyId), nil)
	if err != nil {
		return nil, err
	}

	key := Key{}
	id, _, err := decodeSingle(body, &key)
	if err != nil {
		return nil, err
	}
	key.Id = id

	return &key, nil
}

// CreateKey adds an SSH key to a server. The Forge API accepts the request
// asynchronously and does not return the key, so it is looked up by name.
func (c *Client) CreateKey(serverId string, req *KeyCreateRequest) (*Key, error) {
	payload := map[string]interface{}{
		"name": req.Name,
		"key":  req.Key,
	}
	if req.Username != "" {
		payload["user"] = req.Username
	}

	body, err := c.do(http.MethodPost, c.orgURL("/servers/%s/ssh-keys", serverId), payload)

	if err != nil && req.Overwrite && isNameTaken(err) {
		if err := c.deleteKeyByName(serverId, req.Name); err != nil {
			return nil, err
		}

		body, err = c.do(http.MethodPost, c.orgURL("/servers/%s/ssh-keys", serverId), payload)
	}

	if err != nil {
		return nil, err
	}

	key := Key{}
	id, _, err := decodeSingle(body, &key)
	if err != nil {
		return nil, err
	}

	if id != "" {
		key.Id = id

		return &key, nil
	}

	var found *Key
	err = WaitUntil(2*time.Minute, func() (bool, error) {
		var err error
		found, err = c.SearchKeyByName(serverId, req.Name)

		return found != nil, err
	})
	if err != nil {
		return nil, err
	}

	return found, nil
}

func isNameTaken(err error) bool {
	apiErr, ok := err.(*APIError)

	return ok && apiErr.StatusCode == http.StatusUnprocessableEntity && strings.Contains(apiErr.Body, "already been taken")
}

func (c *Client) deleteKeyByName(serverId string, name string) error {
	key, err := c.SearchKeyByName(serverId, name)
	if err != nil || key == nil {
		return err
	}

	if err := c.DeleteKey(serverId, key.Id); err != nil && !IsNotFound(err) {
		return err
	}

	// The deletion is asynchronous, so wait for the name to be released.
	return WaitUntil(2*time.Minute, func() (bool, error) {
		found, err := c.SearchKeyByName(serverId, name)

		return found == nil, err
	})
}

func (c *Client) DeleteKey(serverId string, keyId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/ssh-keys/%s", serverId, keyId), nil)

	return err
}

func (c *Client) ListKeys(serverId string) ([]Key, error) {
	items, err := c.list(c.orgURL("/servers/%s/ssh-keys", serverId), nil)
	if err != nil {
		return nil, err
	}

	keys := make([]Key, 0, len(items))
	for _, item := range items {
		key := Key{}
		if err := unmarshalAttributes(item, &key); err != nil {
			return nil, err
		}
		key.Id = item.ID
		keys = append(keys, key)
	}

	return keys, nil
}

func (c *Client) SearchKeyByName(serverId string, keyName string) (*Key, error) {
	keys, err := c.ListKeys(serverId)
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Name == keyName {
			return &key, nil
		}
	}

	return nil, nil
}
