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

func (c *Client) GetRedirectRule(serverId string, siteId string, redirectRuleId string) (*RedirectRule, diag.Diagnostics) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/sites/%s/redirect-rules/%s", c.HostURL, serverId, siteId, redirectRuleId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetRedirectRule] RuleId: %s", redirectRuleId)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	redirectRule := RedirectRuleResponse{}
	err = json.Unmarshal(body, &redirectRule)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}
	log.Printf("[INFO] [LARAVELFORGE:GetRedirectRule] Rule: %#v, Body: %#v", &redirectRule, body)

	return &redirectRule.RedirectRule, nil
}

func (c *Client) CreateRedirectRule(serverId string, siteId string, createRuleRequest *CreateRedirectRuleRequest) (*RedirectRule, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateRedirectRule]")
	rb, err := json.Marshal(createRuleRequest)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/sites/%s/redirect-rules", c.HostURL, serverId, siteId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	redirectRule := RedirectRuleResponse{}
	err = json.Unmarshal(body, &redirectRule)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &redirectRule.RedirectRule, nil
}

func (c *Client) DeleteRedirectRule(serverId string, siteId string, redirectRuleId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/sites/%s/redirect-rules/%s", c.HostURL, serverId, siteId, redirectRuleId), nil)
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
