package api

import (
	"fmt"
	"net/url"
)

// PowerOperation stores the attributes for a Power operation
type PowerOperation struct {
	Nodes  []string `json:"nodes"`
	Status string   `json:"status"`
	Level  string   `json:"level,omitempty"`
}

// RebootOperation stores the attributes for a Reboot operation
type RebootOperation struct {
	Kind  string   `json:"kind"`
	Nodes []string `json:"nodes"`
	Level string   `json:"level,omitempty"`
}

// OperationResponse stores the attributes of the response of the submission of an operation
type OperationResponse struct {
	WID string `json:"wid"`
}

// OperationWorkflow stores the attributes of the Workflow of an operation
type OperationWorkflow struct {
	WID   string              `json:"wid"`
	Done  bool                `json:"done"`
	Error bool                `json:"error"`
	Nodes map[string][]string `json:"nodes"` // possible keys: ok, ko, processing
}

// OperationStates stores the State attributes of each nodes concerned by a workflow
type OperationStates map[string]struct {
	Macro string `json:"macro"`
	Micro string `json:"micro"`
	State string `json:"state"`
	Out   string `json:"out,omitempty"`
}

// DeploymentRequest represents a new deployment submission
type DeploymentRequest struct {
	Nodes       []string `json:"nodes"`
	Environment string   `json:"environment"`
	Key         string   `json:"key"`
}

// Deployment represents the response of a new deployment request
type DeploymentResponse struct {
	UID string `json:"uid"`
}

// SubmitPowerOperation submit a power operation to the Kadeploy3 API
func (c *Client) SubmitPowerOperation(operation PowerOperation) (*OperationResponse, error) {
	// send power operation to kadeploy3 API
	req, err := c.caller.R().
		SetBody(operation).
		SetResult(&OperationResponse{}).
		Put(c.getEndpoint("internal/kadeployapi", "/power", url.Values{}))

	if err != nil {
		return nil, fmt.Errorf("Error while sending the power operation: '%s'", err)
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after sending the power operation: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	res, ok := req.Result().(*OperationResponse)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the reboot submission (unexpected type)")
	}

	return res, nil
}

// RequestPowerStatus request the power status of the node to the Kadeploy3 API
func (c *Client) RequestPowerStatus(node string) (*OperationResponse, error) {
	// send power operation to kadeploy3 API
	req, err := c.caller.R().
		SetResult(&OperationResponse{}).
		Get(c.getEndpoint("internal/kadeployapi", "/power", url.Values{"nodes": []string{node}}))

	if err != nil {
		return nil, err
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after sending the power operation: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	res, ok := req.Result().(*OperationResponse)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the reboot submission (unexpected type)")
	}

	return res, nil
}

// SubmitRebootOperation submit a reboot operation to the Kadeploy3 API
func (c *Client) SubmitRebootOperation(operation RebootOperation) (*OperationResponse, error) {
	// send reboot operation to kadeploy3 API
	req, err := c.caller.R().
		SetBody(operation).
		SetResult(&OperationResponse{}).
		Post(c.getEndpoint("internal/kadeployapi", "/reboot", url.Values{}))

	if err != nil {
		return nil, fmt.Errorf("Error while sending the reboot operation: '%s'", err)
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after sending the reboot operation: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	res, ok := req.Result().(*OperationResponse)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the reboot submission (unexpected type)")
	}

	return res, nil
}

// SubmitDeployment submits a new deployment request to g5k api
func (c *Client) SubmitDeployment(operation DeploymentRequest) (*DeploymentResponse, error) {
	// send deployment request to kadeploy3 API
	req, err := c.caller.R().
		SetBody(operation).
		SetResult(&DeploymentResponse{}).
		Post(c.getEndpoint("deployments", "/", url.Values{}))

	if err != nil {
		return nil, fmt.Errorf("Error while sending the deployment request: '%s'", err)
	}

	// check HTTP error code (expected: 201 OK)
	if req.StatusCode() != 201 {
		return nil, fmt.Errorf("The server returned an error (code: %d) after sending Deployment request: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	res, ok := req.Result().(*DeploymentResponse)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the Deployment request (unexpected type)")
	}

	return res, nil
}

// GetOperationWorkflow fetch and return an operation workflow from its ID
func (c *Client) GetOperationWorkflow(operation string, wid string) (*OperationWorkflow, error) {
	// get workflow fron kadeploy3 API
	req, err := c.caller.R().
		SetResult(&OperationWorkflow{}).
		Get(c.getEndpoint("internal/kadeployapi", fmt.Sprintf("/%s/%s", operation, wid), url.Values{}))

	if err != nil {
		return nil, err
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) while fetching the operation workflow: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	workflow, ok := req.Result().(*OperationWorkflow)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the operation workflow (unexpected type)")
	}

	return workflow, nil
}

// GetOperationStates fetch and return the states of an operation workflow from its ID
func (c *Client) GetOperationStates(operation string, wid string) (*OperationStates, error) {
	// get workflow fron kadeploy3 API
	req, err := c.caller.R().
		SetResult(&OperationStates{}).
		Get(c.getEndpoint("internal/kadeployapi", fmt.Sprintf("/%s/%s/state", operation, wid), url.Values{}))

	if err != nil {
		return nil, err
	}

	// check HTTP error code (expected: 200 OK)
	if req.StatusCode() != 200 {
		return nil, fmt.Errorf("The server returned an error (code: %d) while fetching the operation states: '%s'", req.StatusCode(), req.Status())
	}

	// unmarshal result
	states, ok := req.Result().(*OperationStates)
	if !ok {
		return nil, fmt.Errorf("Error in the response of the operation states (unexpected type)")
	}

	return states, nil
}
