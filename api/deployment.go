package api

import (
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
}

// SubmitDeployment submits a new deployment request to g5k api
func (c *Client) SubmitDeployment(deploymentReq DeploymentRequest) (string, error) {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/deployments", G5kAPIFrontend, c.Site)

	log.Infof("Submitting a new deployment... (image: '%s')", deploymentReq.Environment)

	// send deployment request
	deploymentRes, err := c.Request().
		SetHeader("Content-Type", "application/json").
		SetBody(deploymentReq).
		SetResult(&Deployment{}).
		Post(url)

	if err != nil {
		return "", fmt.Errorf("Error while sending the deployment request: '%s'", err)
	}

	// unmarshal result
	deployment, ok := deploymentRes.Result().(*Deployment)
	if !ok {
		return "", fmt.Errorf("Error in the response of the Deployment request (unexpected type)")
	}

	log.Infof("Deployment submitted successfully (id: '%s')", deployment.UID)
	return deployment.UID, nil
}

// GetDeployment get the deployment from its id
func (c *Client) GetDeployment(deploymentID string) (*Deployment, error) {
	// create url for API call
	url := fmt.Sprintf("%s/sites/%s/deployments/%s", G5kAPIFrontend, c.Site, deploymentID)

	// send request
	deploymentRes, err := c.Request().
		SetResult(&Deployment{}).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("Error while retrieving Deployment informations: '%s'", err)
	}

	deployment, ok := deploymentRes.Result().(*Deployment)
	if !ok {
		return nil, fmt.Errorf("Error in the Deployment retrieving (unexpected type)")
	}

	return deployment, nil
}

// WaitUntilDeploymentIsFinished will wait until the deployment reach the 'terminated' state (no timeout)
func (c *Client) WaitUntilDeploymentIsFinished(deploymentID string) error {
	log.Info("Waiting for deployment to finish, it will take a few minutes...")

	// refresh deployment status
	for deployment, err := c.GetDeployment(deploymentID); deployment.Status != "terminated"; deployment, err = c.GetDeployment(deploymentID) {
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
