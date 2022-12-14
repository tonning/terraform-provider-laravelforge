package client

// User -
type User struct {
	ID int `json:"id,omitempty"`
}

type Server struct {
	ID               int    `json:"id,omitempty"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	PhpVersion       string `json:"php_version"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
	IsReady          bool   `json:"is_ready"`
	//CredentialId     int    `json:"credential_id"`
	//Size             string `json:"size"`
	//Region           string `json:"region"`
	//DatabaseType     string `json:"database_type"`
	//CreatedAt        string `json:"created_at"`
}

type ServerResponse struct {
	Server           Server `json:"server"`
	ProvisionCommand string `json:"provision_command"`
	SudoPassword     string `json:"sudo_password"`
}

type ServerCreateRequest struct {
	Name             string `json:"name"`
	Provider         string `json:"provider"`
	Type             string `json:"type"`
	UbuntuVersion    string `json:"ubuntu_version"`
	PhpVersion       string `json:"php_version"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
}

type ServerUpdateRequest struct {
	Name             string `json:"name"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
	//MaxUploadSize    int    `json:"max_upload_size"`
	//Timezone         string `json:"timezone"`
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

type Key struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type KeyGet struct {
	Key Key `json:"key"`
}

type KeyCreateRequest struct {
	Name     string `json:"name"`
	Key      string `json:"key"`
	Username string `json:"username"`
}
