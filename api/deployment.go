package api

import (
	"fmt"
	"sort"

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

	// check HTTP error code (expected: 201 Created)
	if deploymentRes.StatusCode() != 201 {
		return "", fmt.Errorf("The server returned an error (code: %d) after sending Deployment request: '%s'", deploymentRes.StatusCode(), deploymentRes.Status())
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

	// check HTTP error code (expected: 200 OK)
	if deploymentRes.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after requesting Deployment informations: '%s'", deploymentRes.StatusCode(), deploymentRes.Status())
	}

	// unmarshal result
	deployment, ok := deploymentRes.Result().(*Deployment)
	if !ok {
		return nil, fmt.Errorf("Error in the Deployment retrieving (unexpected type)")
	}

	sort.Strings(deployment.Nodes)
	return deployment, nil
}
