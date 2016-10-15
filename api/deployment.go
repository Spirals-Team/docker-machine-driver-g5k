package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/machine/libmachine/log"
)

// DeploymentRequest represents a new deployment request
type DeploymentRequest struct {
	Nodes       []string `json:"nodes"`
	Environment string   `json:"environment"`
	Key         string   `json:"key"`
}

// Deployment represents a deployment response
type Deployment struct {
	Nodes  []string `json:"nodes"`
	Site   string   `json:"site_uid"`
	Status string   `json:"status"`
	UID    string   `json:"uid"`
	Links  []Link   `json:"links"`
}

// SubmitDeployment submits a new deployment request to g5k api
func (a *Api) SubmitDeployment(deploymentReq DeploymentRequest) (string, error) {
	// create url for API call
	urlDeploy := fmt.Sprintf("%s/sites/%s/deployments", G5kApiFrontend, a.Site)

	// create deployment request json
	deploymentArguments, err := json.Marshal(deploymentReq)
	if err != nil {
		return "", err
	}

	log.Infof("Submitting a new deployment... (image: '%s')", deploymentReq.Environment)

	// send deployment request
	resp, err := a.post(urlDeploy, string(deploymentArguments))
	if err != nil {
		return "", err
	}

	// unmarshal deployment response
	var deployment Deployment
	err = json.Unmarshal(resp, &deployment)
	if err != nil {
		return "", err
	}

	log.Infof("Deployment submitted successfully (id: '%s')", deployment.UID)
	return deployment.UID, nil
}

// GetDeployment get the deployment from its id
func (a *Api) GetDeployment(deploymentID string) (*Deployment, error) {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/deployments/%s", G5kApiFrontend, a.Site, deploymentID)

	// send request
	resp, err := a.get(url)
	if err != nil {
		return nil, err
	}

	// unmarshal json response
	var deployment Deployment
	err = json.Unmarshal(resp, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

// WaitUntilDeploymentIsFinished will wait until the deployment reach the 'terminated' state (no timeout)
func (a *Api) WaitUntilDeploymentIsFinished(deploymentID string) error {
	log.Info("Waiting for deployment to finish, it will take a few minutes...")

	// refresh deployment status
	for deployment, err := a.GetDeployment(deploymentID); deployment.Status != "terminated"; deployment, err = a.GetDeployment(deploymentID) {
		// check if GetDeployment returned an error
		if err != nil {
			return err
		}

		// stop if the deployment is in 'canceled' or 'error' state
		if deployment.Status == "canceled" || deployment.Status == "error" {
			return fmt.Errorf("Can't wait for a deployment in '%s' state", deployment.Status)
		}

		// wait 10 seconds before making another API call
		time.Sleep(10 * time.Second)
	}

	log.Info("Deployment finished successfully")
	return nil
}
