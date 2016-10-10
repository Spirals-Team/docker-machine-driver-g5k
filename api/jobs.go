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

	log.Info("Job submitted successfully (id: %s)", job.UID)
	return job.UID, err
}

// Refresh job's state
func (a *Api) GetJob(jobId int) (*Job, error) {
	job := new(Job)
	url := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kApiFrontend, a.Site, jobId)

	if resp, err := a.get(url); err != nil {
		return &Job{}, err
	} else {
		err = json.Unmarshal(resp, &job)
		return job, err
	}
}

// Returns the job's current state
func (a *Api) GetJobState(jobId int) (string, error) {
	if job, err := a.GetJob(jobId); err != nil {
		return "", err
	} else if a.jobIsOver(job) {
		return "terminated", nil
	} else {
		return job.State, nil
	}
}

// Returns true if the job expired, false otherwise
func (a *Api) jobIsOver(job *Job) bool {
	currentTime := time.Now().Unix()
	startTime := int64(job.StartTime)
	timelife := int64(job.Timelife)

	return (currentTime - startTime) >= timelife
}

// Free the nodes allocated to the jobs
func (a *Api) KillJob(jobId int) error {
	url := fmt.Sprintf("%s/sites/%s/jobs/%v", G5kApiFrontend, a.Site, jobId)

	_, err := a.del(url)

	return err
}

func (a *Api) waitJobIsReady(job *Job) bool {
	var err error
	tmp_job := new(Job)

	for job.State == "waiting" || job.State == "tolaunch" || job.State == "launching" {
		if tmp_job, err = a.GetJob(job.UID); err != nil {
			return false
		}
		*job = *tmp_job
		time.Sleep(3 * time.Second)
	}

	// If the launching failed
	if job.State != "running" {
		return false
	} else {
		return true
	}
}
