package client

import (
	"fmt"
	"net/http"
)

type Certificate struct {
	Id            string `json:"-"`
	DomainId      string `json:"-"`
	Type          string `json:"type"`
	RequestStatus string `json:"request_status"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	Active        bool   `json:"active"`
}

// CertificateCreateRequest describes a certificate for a single domain of a site.
type CertificateCreateRequest struct {
	Type               string
	Enable             bool
	VerificationMethod string
	ExistingKey        string
	ExistingCert       string
	CloneCertificateId int
}

func (r *CertificateCreateRequest) payload() map[string]interface{} {
	p := map[string]interface{}{
		"type":   r.Type,
		"enable": r.Enable,
	}

	switch r.Type {
	case "letsencrypt":
		p["letsencrypt"] = map[string]interface{}{"verification_method": r.VerificationMethod}
	case "existing":
		p["existing"] = map[string]interface{}{"key": r.ExistingKey, "certificate": r.ExistingCert}
	case "clone":
		p["clone"] = map[string]interface{}{"certificate_id": r.CloneCertificateId}
	}

	return p
}

func (c *Client) certificatesURL(serverId, siteId, domainId string) string {
	return c.orgURL("/servers/%s/sites/%s/domains/%s/certificates", serverId, siteId, domainId)
}

func decodeCertificate(body []byte, domainId string) (*Certificate, error) {
	certificate := Certificate{}
	id, _, err := decodeSingle(body, &certificate)
	if err != nil {
		return nil, err
	}
	certificate.Id = id
	certificate.DomainId = domainId

	return &certificate, nil
}

func (c *Client) GetCertificate(serverId string, siteId string, domainId string, certificateId string) (*Certificate, error) {
	body, err := c.do(http.MethodGet, c.certificatesURL(serverId, siteId, domainId)+"/"+certificateId, nil)
	if err != nil {
		return nil, err
	}

	return decodeCertificate(body, domainId)
}

// FindCertificate locates a certificate by id across all domains of the site.
func (c *Client) FindCertificate(serverId string, siteId string, certificateId string) (*Certificate, error) {
	domains, err := c.ListDomains(serverId, siteId)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		certificate, err := c.GetCertificate(serverId, siteId, domain.Id, certificateId)
		if err == nil {
			return certificate, nil
		}
		if !IsNotFound(err) {
			return nil, err
		}
	}

	return nil, &APIError{StatusCode: http.StatusNotFound, Body: fmt.Sprintf("certificate %s not found", certificateId)}
}

func (c *Client) CreateCertificate(serverId string, siteId string, domainId string, request *CertificateCreateRequest) (*Certificate, error) {
	body, err := c.do(http.MethodPost, c.certificatesURL(serverId, siteId, domainId), request.payload())
	if err != nil {
		return nil, err
	}

	return decodeCertificate(body, domainId)
}

func (c *Client) EnableCertificate(serverId string, siteId string, domainId string, certificateId string) error {
	_, err := c.do(http.MethodPost, c.certificatesURL(serverId, siteId, domainId)+"/"+certificateId+"/actions", map[string]string{"action": "enable"})

	return err
}

func (c *Client) DeleteCertificate(serverId string, siteId string, domainId string, certificateId string) error {
	_, err := c.do(http.MethodDelete, c.certificatesURL(serverId, siteId, domainId)+"/"+certificateId, nil)

	return err
}
