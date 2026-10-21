package client

import (
	"net/http"
)

type ScheduledJob struct {
	Id        string
	Name      string `json:"name"`
	Command   string `json:"command"`
	User      string `json:"user"`
	Frequency string `json:"frequency"`
	Cron      string `json:"cron"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type CreateScheduledJob struct {
	Command   string `json:"command"`
	Frequency string `json:"frequency"`
	User      string `json:"user"`
	Cron      string `json:"cron,omitempty"`
}

func (c *Client) GetScheduledJob(serverId string, jobId string) (*ScheduledJob, error) {
	body, err := c.do(http.MethodGet, c.orgURL("/servers/%s/scheduled-jobs/%s", serverId, jobId), nil)
	if err != nil {
		return nil, err
	}

	return decodeScheduledJob(body)
}

func (c *Client) CreateScheduledJob(serverId string, createJob *CreateScheduledJob) (*ScheduledJob, error) {
	body, err := c.do(http.MethodPost, c.orgURL("/servers/%s/scheduled-jobs", serverId), createJob)
	if err != nil {
		return nil, err
	}

	return decodeScheduledJob(body)
}

func (c *Client) DeleteScheduledJob(serverId string, jobId string) error {
	_, err := c.do(http.MethodDelete, c.orgURL("/servers/%s/scheduled-jobs/%s", serverId, jobId), nil)

	return err
}

func decodeScheduledJob(body []byte) (*ScheduledJob, error) {
	job := ScheduledJob{}
	id, _, err := decodeSingle(body, &job)
	if err != nil {
		return nil, err
	}
	job.Id = id

	return &job, nil
}
