package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetUser(userID string) (*User, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/users/%s", c.HostURL, userID), nil)
	if err != nil {
		return nil, err
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	user := User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
