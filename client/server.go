package client

import (
	"encoding/json"
	"net/http"
)

type Server struct {
	Id               string     `json:"-"`
	CredentialId     FlexString `json:"credential_id"`
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Provider         string     `json:"provider"`
	Identifier       string     `json:"identifier"`
	Size             string     `json:"size"`
	Region           string     `json:"region"`
	UbuntuVersion    string     `json:"ubuntu_version"`
	DbStatus         string     `json:"db_status"`
	RedisStatus      string     `json:"redis_status"`
	OpcacheStatus    string     `json:"opcache_status"`
	PhpVersion       string     `json:"php_version"`
	PhpCliVersion    string     `json:"php_cli_version"`
	DatabaseType     string     `json:"database_type"`
	IpAddress        string     `json:"ip_address"`
	SshPort          int        `json:"ssh_port"`
	PrivateIpAddress string     `json:"private_ip_address"`
	LocalPublicKey   string     `json:"local_public_key"`
	Revoked          bool       `json:"revoked"`
	CreatedAt        string     `json:"created_at"`
	IsReady          bool       `json:"is_ready"`

	// Only present in the response of a server creation.
	ProvisionCommand string `json:"provision_command"`
	SudoPassword     string `json:"sudo_password"`
}

type ServerCreateRequest struct {
	Name             string
	Provider         string
	CredentialId     int
	Type             string
	UbuntuVersion    string
	PhpVersion       string
	DatabaseType     string
	Region           string
	Size             string
	VpcUuid          string
	IpAddress        string
	PrivateIpAddress string
	SshPort          int
}

type ServerUpdateRequest struct {
	Name             string `json:"name,omitempty"`
	IpAddress        string `json:"ip_address,omitempty"`
	PrivateIpAddress string `json:"private_ip_address,omitempty"`
}

func (r *ServerCreateRequest) payload() map[string]interface{} {
	p := map[string]interface{}{
		"name":     r.Name,
		"provider": r.Provider,
		"type":     r.Type,
	}

	if r.CredentialId != 0 {
		p["credential_id"] = r.CredentialId
	}
	if r.UbuntuVersion != "" {
		p["ubuntu_version"] = r.UbuntuVersion
	}
	if r.PhpVersion != "" {
		p["php_version"] = r.PhpVersion
	}
	if r.DatabaseType != "" {
		p["database_type"] = r.DatabaseType
	}

	if r.Provider == "custom" {
		custom := map[string]interface{}{"ip_address": r.IpAddress}
		if r.PrivateIpAddress != "" {
			custom["private_ip_address"] = r.PrivateIpAddress
		}
		if r.SshPort != 0 {
			custom["ssh_port"] = r.SshPort
		}
		p["custom"] = custom
	} else {
		settings := map[string]interface{}{}
		if r.Region != "" {
			settings["region_id"] = r.Region
		}
		if r.Size != "" {
			settings["size_id"] = r.Size
		}
		if r.VpcUuid != "" {
			settings["vpc_uuid"] = r.VpcUuid
		}
		p[r.Provider] = settings
	}

	return p
}

func decodeServer(body []byte) (*Server, error) {
	server := Server{}
	id, meta, err := decodeSingle(body, &server)
	if err != nil {
		return nil, err
	}
	server.Id = id

	if len(meta) > 0 && (server.ProvisionCommand == "" || server.SudoPassword == "") {
		var m struct {
			ProvisionCommand string `json:"provision_command"`
			SudoPassword     string `json:"sudo_password"`
		}
		if err := json.Unmarshal(meta, &m); err == nil {
			if server.ProvisionCommand == "" {
				server.ProvisionCommand = m.ProvisionCommand
			}
			if server.SudoPassword == "" {
				server.SudoPassword = m.SudoPassword
			}
		}
	}

	return &server, nil
}

func (c *Client) GetServer(serverId string) (*Server, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s", serverId), nil)
	if err != nil {
		return nil, err
	}

	return decodeServer(body)
}

func (c *Client) CreateServer(createServer *ServerCreateRequest) (*Server, error) {
	body, err := c.do(http.MethodPost, c.orgURL("/servers"), createServer.payload())
	if err != nil {
		return nil, err
	}

	return decodeServer(body)
}

func (c *Client) UpdateServer(serverId string, serverUpdates ServerUpdateRequest) error {
	_, err := c.do(http.MethodPut, c.orgURL("/servers/%s", serverId), serverUpdates)

	return err
}

func (c *Client) DeleteServer(serverId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s", serverId), nil)

	return err
}

// GetServerPublicKey returns the public SSH key Forge uses to connect to the server.
func (c *Client) GetServerPublicKey(serverId string) (string, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s/key", serverId), nil)
	if err != nil {
		return "", err
	}

	attrs := struct {
		PublicKey string `json:"public_key"`
	}{}
	if _, _, err := decodeSingle(body, &attrs); err != nil {
		return "", err
	}

	return attrs.PublicKey, nil
}

func (c *Client) EnableOpcache(serverId string) error {
	_, err := c.do(http.MethodPost, c.orgURL("/servers/%s/php/opcache", serverId), nil)

	return err
}

func (c *Client) DisableOpcache(serverId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/php/opcache", serverId), nil)

	return err
}
