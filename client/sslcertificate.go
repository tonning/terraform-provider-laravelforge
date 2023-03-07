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

func (c *Client) GetCertificate(serverId string, siteId string, certificateId string) (*Certificate, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/sites/%s/certificates/%s", c.HostURL, serverId, siteId, certificateId), nil)

	log.Printf("[INFO] [LARAVELFORGE:GetCertificate] Certificate ID: %s", certificateId)
	if err != nil {
		return nil, err
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	certificate := CertificateResponse{}
	err = json.Unmarshal(body, &certificate)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] [LARAVELFORGE:GetCertificate] Certificate: %#v, Body: %#v", &certificate, body)

	return &certificate.Certificate, nil
}

func (c *Client) ObtainLetsEncryptSslCertificate(serverId string, siteId string, createSslCertificate *SslCertificateCreateRequest) (*Certificate, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateSslCertificate]")
	rb, err := json.Marshal(createSslCertificate)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/sites/%s/certificates/letsencrypt", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	log.Printf("[INFO] [LARAVELFORGE:GetCertificate] Certificate request: %#v, rb: %#v", req, strings.NewReader(string(rb)))
	//return nil, nil
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	certificate := CertificateResponse{}
	err = json.Unmarshal(body, &certificate)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &certificate.Certificate, nil
}

func (c *Client) CloneExistingSslCertificate(serverId string, siteId string, request *SslCertificateCloneRequest) (*Certificate, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CloneExistingSslCertificate]")
	rb, err := json.Marshal(request)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/servers/%s/sites/%s/certificates", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	log.Printf("[INFO] [LARAVELFORGE:CloneExistingSslCertificate] Certificate request: %#v, rb: %#v", req, strings.NewReader(string(rb)))

	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	certificate := CertificateResponse{}
	err = json.Unmarshal(body, &certificate)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &certificate.Certificate, nil
}

func (c *Client) InstallExistingSslCertificate(serverId string, siteId string, request *SslCertificateInstallExistingRequest) (*Certificate, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:InstallExistingSslCertificate]")
	rb, err := json.Marshal(request)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/servers/%s/sites/%s/certificates", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	log.Printf("[INFO] [LARAVELFORGE:InstallExistingSslCertificate] Certificate request: %#v, rb: %#v", req, strings.NewReader(string(rb)))

	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	certificate := CertificateResponse{}
	err = json.Unmarshal(body, &certificate)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &certificate.Certificate, nil
}

func (c *Client) ActivateCertificate(serverId string, siteId string, certificateId string) diag.Diagnostics {
	log.Printf("[INFO] [LARAVELFORGE:ActivateCertificate]")

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/sites/%s/certificates/%s/activate", c.HostURL, serverId, siteId, certificateId), nil)
	if err != nil {
		return diag.Errorf("Whoops: %s", err)
	}

	err = c.doRequestEmptyBody(req)
	if err != nil {
		return diag.Errorf("Whoopsy: %s", err)
	}

	return nil
}

//func (c *Client) UpdateSite(serverId string, siteId string, siteUpdates SiteUpdateRequest) (*Site, diag.Diagnostics) {
//	rb, err := json.Marshal(siteUpdates)
//	if err != nil {
//		return nil, diag.Errorf("Whoops: %s", err)
//	}
//	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/servers/%s/sites/%s", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
//	if err != nil {
//		return nil, diag.Errorf("Whoops: %s", err)
//	}
//
//	body, err, _ := c.doRequest(req)
//	if err != nil {
//		return nil, diag.Errorf("Whoops: %s", err)
//	}
//
//	site := SiteGet{}
//	err = json.Unmarshal(body, &site)
//	if err != nil {
//		return nil, diag.Errorf("Whoops: %s", err)
//	}
//
//	return &site.Site, nil
//}

func (c *Client) DeleteCertificate(serverId string, siteId string, certificateId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/sites/%s/certificates/%s", c.HostURL, serverId, siteId, certificateId), nil)
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
