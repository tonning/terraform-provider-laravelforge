package client

import (
	"net/http"
)

type Daemon struct {
	Id        string
	Command   string `json:"command"`
	User      string `json:"user"`
	Directory string `json:"directory"`
	Processes int    `json:"processes"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CreateDaemonRequest struct {
	Name         string `json:"name"`
	Command      string `json:"command"`
	User         string `json:"user"`
	Directory    string `json:"directory,omitempty"`
	Processes    int    `json:"processes"`
	Startsecs    int    `json:"startsecs"`
	Stopwaitsecs int    `json:"stopwaitsecs"`
	Stopsignal   string `json:"stopsignal,omitempty"`
}

// Daemons are called background processes in the current Forge API.
func (c *Client) GetDaemon(serverId string, daemonId string) (*Daemon, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s/background-processes/%s", serverId, daemonId), nil)
	if err != nil {
		return nil, err
	}

	return decodeDaemon(body)
}

func (c *Client) CreateDaemon(serverId string, createDaemon *CreateDaemonRequest) (*Daemon, error) {
	body, err := c.do(http.MethodPost, c.orgURL("/servers/%s/background-processes", serverId), createDaemon)
	if err != nil {
		return nil, err
	}

	return decodeDaemon(body)
}

func (c *Client) DeleteDaemon(serverId string, daemonId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/background-processes/%s", serverId, daemonId), nil)

	return err
}

func decodeDaemon(body []byte) (*Daemon, error) {
	daemon := Daemon{}
	id, _, err := decodeSingle(body, &daemon)
	if err != nil {
		return nil, err
	}
	daemon.Id = id

	return &daemon, nil
}
