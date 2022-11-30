package client

// User -
type User struct {
	ID int `json:"id,omitempty"`
}

type Server struct {
	ID               int    `json:"id,omitempty"`
	CredentialId     int    `json:"credential_id"`
	Name             string `json:"name"`
	Size             string `json:"size"`
	Region           string `json:"region"`
	PhpVersion       string `json:"php_version"`
	DatabaseType     string `json:"database_type"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
	CreatedAt        string `json:"created_at"`
	IsReady          bool   `json:"is_ready"`
}

type SiteGet struct {
	Site Site `json:"site"`
}

type Site struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Username    string   `json:"username"`
	Directory   string   `json:"directory"`
	Wildcards   bool     `json:"wildcards"`
	Status      string   `json:"status"`
	ProjectType string   `json:"project_type"`
	CreatedAt   string   `json:"created_at"`
	Network     []string `json:"network"`
}

type SiteCreateRequest struct {
	Domain      string `json:"domain"`
	ProjectType string `json:"project_type"`
	Directory   string `json:"directory"`
	Username    string `json:"username"`
	PhpVersion  string `json:"php_version"`
}

type SiteUpdateRequest struct {
	Directory  string `json:"directory"`
	Name       string `json:"name"`
	PhpVersion string `json:"php_version"`
}

type SiteUpdatePhpVersion struct {
	Version string `json:"version"`
}

type ServerCreateRequest struct {
	Provider string `json:"provider"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	//Size             string `json:"size"`
	UbuntuVersion    string `json:"ubuntu_version"`
	PhpVersion       string `json:"php_version"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
}

type SiteItem struct {
	Provider         string `json:"provider"`
	Type             string `json:"type"`
	Name             string `json:"name"`
	Size             string `json:"size"`
	UbuntuVersion    string `json:"ubuntu_version"`
	PhpVersion       string `json:"php_version"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
}
