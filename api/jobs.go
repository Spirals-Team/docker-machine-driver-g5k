package api

import (
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/log"
)

// JobRequest represents a new job submission
type JobRequest struct {
	Resources  string   `json:"resources"`
	Command    string   `json:"command"`
	Properties string   `json:"properties,omitempty"`
	Types      []string `json:"types"`
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

	// unmarshal result
	job, ok := jobRes.Result().(*Job)
	if !ok {
		return nil, fmt.Errorf("Error in the Job retrieving (unexpected type)")
	}

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
	if _, err := c.Request().Delete(url); err != nil {
		return fmt.Errorf("Error while killing job: '%s'", err)
	}

	return nil
}

// WaitUntilJobIsReady wait until the job reach the 'running' state (no timeout)
func (c *Client) WaitUntilJobIsReady(jobID int) error {
	log.Info("Waiting for job to run...")

	// refresh job state
	for job, err := c.GetJob(jobID); job.State != "running"; job, err = c.GetJob(jobID) {
		// check if GetJob returned an error
		if err != nil {
			return err
		}

		// stop if the job is in 'error' or 'terminated' state
		if job.State == "error" || job.State == "terminated" {
			return fmt.Errorf("Can't wait for a job in '%s' state", job.State)
		}

		// warn if job is in 'hold' state
		if job.State == "hold" {
			log.Infof("Job '%s' is in hold state, dont forget to resume it")
		}

		// wait 3 seconds before making another API call
		time.Sleep(3 * time.Second)
	}

	log.Info("Job is running")
	return nil
}
