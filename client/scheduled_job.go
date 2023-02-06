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

func (c *Client) GetScheduledJob(serverId string, jobId string) (*ScheduledJob, diag.Diagnostics) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/servers/%s/jobs/%s", c.HostURL, serverId, jobId), nil)
	log.Printf("[INFO] [LARAVELFORGE:GetScheduledJob] SiteId: %s", jobId)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	job := ScheduledJobResponse{}
	err = json.Unmarshal(body, &job)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}
	log.Printf("[INFO] [LARAVELFORGE:GetScheduledJob] Job: %#v, Body: %#v", &job, body)

	return &job.Job, nil
}

func (c *Client) CreateScheduledJob(serverId string, createJob *CreateScheduledJob) (*ScheduledJob, diag.Diagnostics) {
	log.Printf("[INFO] [LARAVELFORGE:CreateScheduledJob]")
	rb, err := json.Marshal(createJob)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/servers/%s/jobs", c.HostURL, serverId), strings.NewReader(string(rb)))
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	body, err, _ := c.doRequest(req)
	if err != nil {
		return nil, diag.Errorf("Whoopsy: %s", err)
	}

	job := ScheduledJobResponse{}
	err = json.Unmarshal(body, &job)
	if err != nil {
		return nil, diag.Errorf("Whoops: %s", err)
	}

	return &job.Job, nil
}

func (c *Client) DeleteScheduledJob(serverId string, jobId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/servers/%s/jobs/%s", c.HostURL, serverId, jobId), nil)
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
