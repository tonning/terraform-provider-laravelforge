package client

// User -
type User struct {
	ID int `json:"id,omitempty"`
}

type Server struct {
	Id               int    `json:"id"`
	CredentialId     string `json:"credential_id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Provider         string `json:"provider"`
	ProviderId       string `json:"provider_id"`
	Size             string `json:"size"`
	Region           string `json:"region"`
	UbuntuVersion    string `json:"ubuntu_version"`
	DbStatus         string `json:"db_status"`
	RedisStatus      string `json:"redis_status"`
	PhpVersion       string `json:"php_version"`
	PhpCliVersion    string `json:"php_cli_version"`
	DatabaseType     string `json:"database_type"`
	IpAddress        string `json:"ip_address"`
	SshPort          int    `json:"ssh_port"`
	PrivateIpAddress string `json:"private_ip_address"`
	LocalPublicKey   string `json:"local_public_key"`
	BlackfireStatus  string `json:"blackfire_status"`
	PapertrailStatus string `json:"papertrail_status"`
	Revoked          bool   `json:"revoked"`
	CreatedAt        string `json:"created_at"`
	IsReady          bool   `json:"is_ready"`
	//Tags             []interface{} `json:"tags"`
	//PhpVersions      []struct {
	//	Id                 int    `json:"id"`
	//	Version            string `json:"version"`
	//	Status             string `json:"status"`
	//	DisplayableVersion string `json:"displayable_version"`
	//	BinaryName         string `json:"binary_name"`
	//} `json:"php_versions"`
	//Network []int `json:"network"`
}

type ServerResponse struct {
	Server           Server `json:"server"`
	ProvisionCommand string `json:"provision_command"`
	SudoPassword     string `json:"sudo_password"`
}

type ServerCreateRequest struct {
	Name             string `json:"name"`
	Provider         string `json:"provider"`
	CredentialId     string `json:"credential_id"`
	Type             string `json:"type"`
	Region           string `json:"region"`
	UbuntuVersion    string `json:"ubuntu_version"`
	PhpVersion       string `json:"php_version"`
	IpAddress        string `json:"ip_address"`
	PrivateIpAddress string `json:"private_ip_address"`
	Ocean2VpcUuid    string `json:"ocean2_vpc_uuid"`
	Network          []int  `json:"network"`
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
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	Username    string        `json:"username"`
	Directory   string        `json:"directory"`
	Wildcards   bool          `json:"wildcards"`
	Status      string        `json:"status"`
	ProjectType string        `json:"project_type"`
	CreatedAt   string        `json:"created_at"`
	Network     []interface{} `json:"network"`
}

type SiteCreateRequest struct {
	Domain      string `json:"domain"`
	ProjectType string `json:"project_type"`
	Directory   string `json:"directory"`
	Username    string `json:"username"`
	PhpVersion  string `json:"php_version"`
}

type SiteUpdateRequest struct {
	Directory  string        `json:"directory"`
	Name       string        `json:"name"`
	PhpVersion string        `json:"php_version"`
	Wildcards  bool          `json:"wildcards"`
	Aliases    []interface{} `json:"aliases"`
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

type SslCertificateCloneRequest struct {
	Type          string `json:"type"`
	CertificateId int    `json:"certificate_id"`
}

type SslCertificateCreateRequest struct {
	Domains     []interface{} `json:"domains"`
	DnsProvider DnsProvider   `json:"dns_provider"`
}

type DnsProvider struct {
	Type                  string `json:"type"`
	CloudflareApiToken    string `json:"cloudflare_api_token"`
	Route53Key            string `json:"route53_key"`
	Route53Secret         string `json:"route53_secret"`
	DigitaloceanToken     string `json:"digitalocean_token"`
	DnssimpleToken        string `json:"dnssimple_token"`
	LinodeToken           string `json:"linode_token"`
	OvhEndpoint           string `json:"ovh_endpoint"`
	OvhAppKey             string `json:"ovh_app_key"`
	OvhAppSecret          string `json:"ovh_app_secret"`
	OvhConsumerKey        string `json:"ovh_consumer_key"`
	GoogleCredentialsFile string `json:"google_credentials_file"`
}

type Certificate struct {
	Domain        string `json:"domain"`
	Type          string `json:"type"`
	RequestStatus string `json:"request_status"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	Id            int    `json:"id"`
	Existing      bool   `json:"existing"`
	Active        bool   `json:"active"`
}

type CertificateResponse struct {
	Certificate Certificate `json:"certificate"`
}

type CreateScheduledJob struct {
	Command   string `json:"command"`
	Frequency string `json:"frequency"`
	User      string `json:"user"`
	Minute    string `json:"minute"`
	Hour      string `json:"hour"`
	Day       string `json:"day"`
	Month     string `json:"month"`
	Weekday   string `json:"weekday"`
}

type ScheduledJob struct {
	Id        int    `json:"id"`
	Command   string `json:"command"`
	User      string `json:"user"`
	Frequency string `json:"frequency"`
	Cron      string `json:"cron"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type ScheduledJobResponse struct {
	Job ScheduledJob `json:"job"`
}

type CreateDaemonRequest struct {
	Command      string `json:"command"`
	User         string `json:"user"`
	Directory    string `json:"directory"`
	Processes    int    `json:"processes"`
	Startsecs    int    `json:"startsecs"`
	Stopwaitsecs int    `json:"stopwaitsecs"`
	Stopsignal   string `json:"stopsignal"`
}

type Daemon struct {
	Id           int    `json:"id"`
	Command      string `json:"command"`
	User         string `json:"user"`
	Directory    string `json:"directory"`
	Processes    int    `json:"processes"`
	Startsecs    int    `json:"startsecs"`
	Stopwaitsecs int    `json:"stopwaitsecs"`
	Stopsignal   string `json:"stopsignal"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
}

type DaemonResponse struct {
	Daemon Daemon `json:"daemon"`
}
