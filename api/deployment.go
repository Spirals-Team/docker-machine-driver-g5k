package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

type Deployment struct {
	Nodes  []string `json:"nodes"`
	Site   string   `json:"site_uid"`
	Status string   `json:"status"`
	UID    string   `json:"uid"`
	Links  []Link   `json:"links"`
}

func (a *Api) DeployEnvironment(jobId int, SSHPublicKeyPath string) error {
	urlDeploy := fmt.Sprintf("%s/sites/%s/deployments", G5kApiFrontend, a.Site)

	job, err := a.GetJob(jobId)
	if err != nil {
		return err
	}

	// read ssh public key
	sshPublicKey, err := a.readSSHPublicKey(SSHPublicKeyPath)
	if err != nil {
		return err
	}

	// Wait for the nodes to be running
	if !a.waitJobIsReady(job) {
		return fmt.Errorf("[G5K_api] Job launching failed")
	}

	// Format arguments
	nodesStrs := make([]string, 0)
	for _, nodes := range job.Nodes {
		nodesStrs = append(nodesStrs, `"`+nodes+`"`)
	}
	nodesJson := strings.Join(nodesStrs, ",")

	// Deploying
	deploymentArguments := fmt.Sprintf(`{"nodes": [%s], "environment": %q, "key": %q}`, nodesJson, defaultImage, sshPublicKey)
	var resp []byte
	var deployment Deployment

	resp, err = a.post(urlDeploy, deploymentArguments)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resp, &deployment)

	// Waiting the deployment finishes
	for deployment.Status == "waiting" || deployment.Status == "processing" {
		time.Sleep(10 * time.Second)
		resp, err = a.get(urlDeploy + "/" + deployment.UID)
		if err != nil {
			return err
		} else if err = json.Unmarshal(resp, &deployment); err != nil {
			return err
		}
	}
	if deployment.Status != "terminated" {
		return fmt.Errorf("[G5K_api] Deployment failed: status is %s\n", deployment.Status)
	}
	return nil
}

// readSSHPublicKey read the ssh public key file and return the key in a string
func (a *Api) readSSHPublicKey(SSHPublicKeyPath string) (string, error) {
	content, err := ioutil.ReadFile(SSHPublicKeyPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
