package api

import (
	"encoding/json"
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
	Links     []Link   `json:"links"`
	Nodes     []string `json:"assigned_nodes"`
}

// SubmitJob submit a new job on g5k api and return the job id
func (a *Api) SubmitJob(jobReq JobRequest) (int, error) {
	// create url for API call
	urlAPI := fmt.Sprintf("%s/sites/%s/jobs", G5kApiFrontend, a.Site)

	// create job request json
	params, err := json.Marshal(jobReq)
	if err != nil {
		return 0, err
	}

	log.Info("Submitting a new job...")

	// send job request
	resp, err := a.post(urlAPI, string(params))
	if err != nil {
		return 0, err
	}

	// unmarshal json response
	var job Job
	err = json.Unmarshal(resp, &job)
	if err != nil {
		return 0, err
	}

	log.Infof("Job submitted successfully (id: '%v')", job.UID)
	return job.UID, nil
}

// GetJob get the job from its id
func (a *Api) GetJob(jobID int) (*Job, error) {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kApiFrontend, a.Site, jobID)

	// send request
	resp, err := a.get(url)
	if err != nil {
		return nil, err
	}

	// unmarshal json response
	var job Job
	err = json.Unmarshal(resp, &job)
	if err != nil {
		return nil, err
	}

	return &job, nil
}

// GetJobState returns the current state of the job
func (a *Api) GetJobState(jobID int) (string, error) {
	// get job from api
	job, err := a.GetJob(jobID)
	if err != nil {
		return "", err
	}

	return job.State, nil
}

// KillJob ask for deletion of a job
func (a *Api) KillJob(jobID int) error {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kApiFrontend, a.Site, jobID)

	// send delete request
	_, err := a.del(url)

	return err
}

// WaitUntilJobIsReady wait until the job reach the 'running' state (no timeout)
func (a *Api) WaitUntilJobIsReady(jobID int) error {
	log.Info("Waiting for job to run...")

	// refresh job state
	for job, err := a.GetJob(jobID); job.State != "running"; job, err = a.GetJob(jobID) {
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
