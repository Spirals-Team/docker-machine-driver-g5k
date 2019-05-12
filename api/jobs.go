package api

import (
	"fmt"
	"sort"

	"github.com/docker/machine/libmachine/log"
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
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/jobs", G5kAPIFrontend, c.Site)

	log.Info("Submitting a new job...")

	// send job request
	jobRes, err := c.Request().
		SetHeader("Content-Type", "application/json").
		SetBody(jobReq).
		SetResult(&Job{}).
		Post(url)

	if err != nil {
		return 0, fmt.Errorf("Error while sending Job submission: '%s'", err)
	}

	// check HTTP error code (expected: 201 Created)
	if jobRes.StatusCode() != 201 {
		return 0, fmt.Errorf("The server returned an error (code: %d) after sending Job submission: '%s'", jobRes.StatusCode(), jobRes.Status())
	}

	// unmarshal result
	job, ok := jobRes.Result().(*Job)
	if !ok {
		return 0, fmt.Errorf("Error in the response of the Job submission (unexpected type)")
	}

	log.Infof("Job submitted successfully (id: '%v')", job.UID)
	return job.UID, nil
}

// GetJob get the job from its id
func (c *Client) GetJob(jobID int) (*Job, error) {
	// create url for API call
	urlJob := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kAPIFrontend, c.Site, jobID)

	// send request
	jobRes, err := c.Request().
		SetResult(&Job{}).
		Get(urlJob)

	if err != nil {
		return nil, fmt.Errorf("Error while retrieving Job informations")
	}

	// check HTTP error code (expected: 200 OK)
	if jobRes.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after requesting Job informations: '%s'", jobRes.StatusCode(), jobRes.Status())
	}

	// unmarshal result
	job, ok := jobRes.Result().(*Job)
	if !ok {
		return nil, fmt.Errorf("Error in the Job retrieving (unexpected type)")
	}

	sort.Strings(job.Types)
	sort.Strings(job.Nodes)
	return job, nil
}

// GetJobState returns the current state of the job
func (c *Client) GetJobState(jobID int) (string, error) {
	// get job from api
	job, err := c.GetJob(jobID)
	if err != nil {
		return "", err
	}

	return job.State, nil
}

// KillJob ask for deletion of a job
func (c *Client) KillJob(jobID int) error {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kAPIFrontend, c.Site, jobID)

	// send delete request
	delRes, err := c.Request().Delete(url)
	if err != nil {
		return fmt.Errorf("Error while killing job: '%s'", err)
	}

	// check HTTP error code (expected: 202 Accepted)
	if delRes.StatusCode() != 202 {
		return fmt.Errorf("The server returned an error (code: %d) after job killing request: '%s'", delRes.StatusCode(), delRes.Status())
	}

	return nil
}
