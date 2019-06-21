package api

import (
	"fmt"
	"net/url"
)

// JobRequest represents a new job submission
type JobRequest struct {
	Resources   string   `json:"resources"`
	Command     string   `json:"command"`
	Properties  string   `json:"properties,omitempty"`
	Reservation string   `json:"reservation,omitempty"`
	Types       []string `json:"types"`
	Queue       string   `json:"queue"`
}

// Job represents an existing job
type Job struct {
	UID       int      `json:"uid"`
	State     string   `json:"state"`
	Timelife  int      `json:"walltime"`
	Types     []string `json:"types"`
	StartTime int      `json:"started_at"`
	Nodes     []string `json:"assigned_nodes"`
}

// SubmitJob submit a new job on g5k api and return the job id
func (c *Client) SubmitJob(jobReq JobRequest) (int, error) {
	// send job request
	req, err := c.getRequest().
		SetHeader("Content-Type", "application/json").
		SetBody(jobReq).
		SetResult(&Job{}).
		Post(c.getEndpoint("jobs", "/", url.Values{}))

	if err != nil {
		return 0, fmt.Errorf("Error while sending Job submission: '%s'", err)
	}

	// check HTTP error code (expected: 201 Created)
	if req.StatusCode() != 201 {
		return 0, fmt.Errorf("The server returned an error (code: %d) after sending Job submission: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	job, ok := req.Result().(*Job)
	if !ok {
		return 0, fmt.Errorf("Error in the response of the Job submission (unexpected type)")
	}

	return job.UID, nil
}

// GetJob get the job from its id
func (c *Client) GetJob(jobID int) (*Job, error) {
	// send request
	req, err := c.getRequest().
		SetResult(&Job{}).
		Get(c.getEndpoint("jobs", fmt.Sprintf("/%v", jobID), url.Values{}))

	if err != nil {
		return nil, fmt.Errorf("Error while retrieving Job informations")
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after requesting Job informations: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	job, ok := req.Result().(*Job)
	if !ok {
		return nil, fmt.Errorf("Error in the Job retrieving (unexpected type)")
	}

	return job, nil
}

// KillJob ask for deletion of a job
func (c *Client) KillJob(jobID int) error {
	// send delete request
	req, err := c.getRequest().Delete(c.getEndpoint("jobs", fmt.Sprintf("/%v", jobID), url.Values{}))
	if err != nil {
		return fmt.Errorf("Error while killing job: '%s'", err)
	}

	// check HTTP error code (202 when accepted or 400 in case the job have already been killed)
	if req.StatusCode() != 202 && req.StatusCode() != 400 {
		return fmt.Errorf("The server returned an error (code: %d) after job killing request: '%s'", req.StatusCode(), req.Status())
	}

	return nil
}
