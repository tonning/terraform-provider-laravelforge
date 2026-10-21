package client

import (
	"net/http"
	"time"
)

type RedirectRule struct {
	Id        string
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CreateRedirectRuleRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
}

func (c *Client) GetRedirectRule(serverId string, siteId string, redirectRuleId string) (*RedirectRule, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s/sites/%s/redirect-rules/%s", serverId, siteId, redirectRuleId), nil)
	if err != nil {
		return nil, err
	}

	rule := RedirectRule{}
	id, _, err := decodeSingle(body, &rule)
	if err != nil {
		return nil, err
	}
	rule.Id = id

	return &rule, nil
}

func (c *Client) ListRedirectRules(serverId string, siteId string) ([]RedirectRule, error) {
	items, err := c.list(c.orgURL("/servers/%s/sites/%s/redirect-rules", serverId, siteId), nil)
	if err != nil {
		return nil, err
	}

	rules := make([]RedirectRule, 0, len(items))
	for _, item := range items {
		rule := RedirectRule{}
		if err := unmarshalAttributes(item, &rule); err != nil {
			return nil, err
		}
		rule.Id = item.ID
		rules = append(rules, rule)
	}

	return rules, nil
}

// CreateRedirectRule creates a rule. The Forge API accepts the request
// asynchronously without returning the rule, so it is looked up afterwards.
func (c *Client) CreateRedirectRule(serverId string, siteId string, createRuleRequest *CreateRedirectRuleRequest) (*RedirectRule, error) {
	body, err := c.do(http.MethodPost, c.orgURL("/servers/%s/sites/%s/redirect-rules", serverId, siteId), createRuleRequest)
	if err != nil {
		return nil, err
	}

	rule := RedirectRule{}
	id, _, err := decodeSingle(body, &rule)
	if err != nil {
		return nil, err
	}
	if id != "" {
		rule.Id = id

		return &rule, nil
	}

	var found *RedirectRule
	err = WaitUntil(2*time.Minute, func() (bool, error) {
		rules, err := c.ListRedirectRules(serverId, siteId)
		if err != nil {
			return false, err
		}

		for _, r := range rules {
			if r.From == createRuleRequest.From && r.To == createRuleRequest.To && r.Type == createRuleRequest.Type {
				r := r
				found = &r

				return true, nil
			}
		}

		return false, nil
	})
	if err != nil {
		return nil, err
	}

	return found, nil
}

func (c *Client) DeleteRedirectRule(serverId string, siteId string, redirectRuleId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/sites/%s/redirect-rules/%s", serverId, siteId, redirectRuleId), nil)

	return err
}
