package driver

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/docker/machine/libmachine/log"
)

// checkVpnConnection check if the VPN is connected and properly configured (DNS) by trying to connect to the site frontend SSH server using its hostname
func (d *Driver) checkVpnConnection() error {
	// construct site frontend hostname
	frontend := fmt.Sprintf("frontend.%s.grid5000.fr:22", d.G5kSite)

	// try to connect to the frontend SSH server
	sshConfig := &gossh.ClientConfig{}
	_, err := gossh.Dial("tcp", frontend, sshConfig)

	// we need to check if the error is network-related because the SSH Dial will always return an error due to the Authentication being not configured
	if _, ok := err.(*net.OpError); ok {
		return fmt.Errorf("Connection to frontend of '%s' site failed. Please check if the site is not undergoing maintenance and your VPN client is connected and properly configured (see driver documentation for more information)", d.G5kSite)
	}

	return nil
}

// generateSSHAuthorizedKeys generate the SSH AuthorizedKeys composed of the driver and user defined key(s)
func (d *Driver) generateSSHAuthorizedKeys() string {
	var authorizedKeysEntries []string

	// add driver key
	authorizedKeysEntries = append(authorizedKeysEntries, "# docker-machine driver g5k - driver key")
	authorizedKeysEntries = append(authorizedKeysEntries, d.DriverSSHPublicKey)

	// add external key(s)
	for index, externalPubKey := range d.ExternalSSHPublicKeys {
		authorizedKeysEntries = append(authorizedKeysEntries, fmt.Sprintf("# docker-machine driver g5k - additional key %d", index))
		authorizedKeysEntries = append(authorizedKeysEntries, strings.TrimSpace(externalPubKey))
	}

	return strings.Join(authorizedKeysEntries, "\n") + "\n"
}
// waitUntilJobIsReady wait until the job reach the 'running' state (no timeout)
func (d *Driver) waitUntilJobIsReady() error {
	log.Info("Waiting for job to run...")

	// refresh job state
	for job, err := d.G5kAPI.GetJob(d.G5kJobID); job.State != "running"; job, err = d.G5kAPI.GetJob(d.G5kJobID) {
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
			log.Infof("Job '%s' is in hold state, dont forget to resume it", d.G5kJobID)
		}

		// wait 3 seconds before making another API call
		time.Sleep(3 * time.Second)
	}

	log.Info("Job is running")
	return nil
}

// makeJobSubmission submit a job submission to Grid'5000
func (d *Driver) makeJobSubmission() error {
	// by default, the node will be redeployed with another image, no specific actions are needed
	jobCommand := "sleep 365d"
	jobTypes := []string{"deploy"}

	// if the user want to reuse the reference environment, specific actions are needed
	if d.G5kReuseRefEnvironment {
		// remove the 'deploy' job type because we will not deploy the machine
		jobTypes = []string{}
		// enable sudo for current user, add public key to ssh authorized keys for root user and wait the end of the job
		jobCommand = `sudo-g5k && echo -n "` + d.generateSSHAuthorizedKeys() + `" |sudo tee -a /root/.ssh/authorized_keys >/dev/null && sleep 365d`
	}

	// submit new Job request
	jobID, err := d.G5kAPI.SubmitJob(api.JobRequest{
		Resources:  fmt.Sprintf("nodes=1,walltime=%s", d.G5kWalltime),
		Command:    jobCommand,
		Properties: d.G5kResourceProperties,
		Types:      jobTypes,
		Queue:      d.G5kJobQueue,
	})
	if err != nil {
		return fmt.Errorf("Error when submitting new job: %s", err.Error())
	}

	d.G5kJobID = jobID
	return nil
}

// makeJobReservation submit a job reservation to Grid'5000
func (d *Driver) makeJobReservation() error {
	jobCommand := "sleep 365d"
	jobTypes := []string{"deploy"}

	// submit new Job request
	jobID, err := d.G5kAPI.SubmitJob(api.JobRequest{
		Resources:   fmt.Sprintf("nodes=1,walltime=%s", d.G5kWalltime),
		Command:     jobCommand,
		Properties:  d.G5kResourceProperties,
		Reservation: d.G5kJobStartTime,
		Types:       jobTypes,
		Queue:       d.G5kJobQueue,
	})
	if err != nil {
		return fmt.Errorf("Error when submitting new job: %s", err.Error())
	}

	d.G5kJobID = jobID
	return nil
}

// waitUntilDeploymentIsFinished will wait until the deployment reach the 'terminated' state (no timeout)
func (d *Driver) waitUntilDeploymentIsFinished(deploymentID string) error {
	log.Info("Waiting for deployment to finish, it will take a few minutes...")

	// refresh deployment status
	for deployment, err := d.G5kAPI.GetDeployment(deploymentID); deployment.Status != "terminated"; deployment, err = d.G5kAPI.GetDeployment(deploymentID) {
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

// handleDeploymentError deallocate the resources when the deployment fail
func (d *Driver) handleDeploymentError() {
	// if deployment fail, we can't recover from this error, so we kill the job
	log.Infof("Unrecoverable error in deployment, killing job ID '%d'...", d.G5kJobID)
	d.G5kAPI.KillJob(d.G5kJobID)
}

// deployImageToNode start the deployment of an OS image to a node
func (d *Driver) deployImageToNode() error {
	// if the user want to reuse Grid'5000 reference environment
	if d.G5kReuseRefEnvironment {
		log.Infof("Skipping host deployment and reusing Grid'5000 standard environment")
		return nil
	}

	// get job informations
	job, err := d.G5kAPI.GetJob(d.G5kJobID)
	if err != nil {
		return fmt.Errorf("Error when getting job (id: '%d') informations: %s", d.G5kJobID, err.Error())
	}

	// check job type before deploying
	if sort.SearchStrings(job.Types, "deploy") != 0 {
		return fmt.Errorf("The job (id: %d) needs to have the type 'deploy'", d.G5kJobID)
	}

	// check if there is only one node for this reservation
	if len(job.Nodes) != 1 {
		return fmt.Errorf("The job (id: '%d') needs to have only one node instead of %d", d.G5kJobID, len(job.Nodes))
	}

	// deploy environment
	deploymentID, err := d.G5kAPI.SubmitDeployment(api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: d.G5kImage,
		Key:         d.generateSSHAuthorizedKeys(),
	})
	if err != nil {
		d.handleDeploymentError()
		return fmt.Errorf("Error when submitting new deployment: %s", err.Error())
	}

	// waiting deployment to finish (REQUIRED or you will interfere with kadeploy)
	if err = d.waitUntilDeploymentIsFinished(deploymentID); err != nil {
		d.handleDeploymentError()
		return fmt.Errorf("Error when waiting for deployment to finish: %s", err.Error())
	}

	return nil
}

// getNodePowerState returns the power status of the node by querying its baseboard management controller (BMC)
func (d *Driver) getNodePowerState() (string, error) {
	node, err := d.GetIP()
	if err != nil {
		return "", fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	op, err := d.G5kAPI.RequestPowerStatus(node)
	if err != nil {
		return "", fmt.Errorf("Failed to request power status: %s", err.Error())
	}

	if err := d.waitUntilWorkflowIsDone("power", op.WID, node); err != nil {
		return "", err
	}

	// get nodes states for the workflow
	states, err := d.G5kAPI.GetOperationStates("power", op.WID)
	if err != nil {
		return "", err
	}

	// get the state of the current node
	state, ok := (*states)[node]
	if !ok {
		return "", fmt.Errorf("Failed to retrieve the workflow state of the power status operation")
	}

	// extract the BMC power status from the state out attribute
	re := regexp.MustCompile(`-bmc: (on|off)$`)
	matches := re.FindStringSubmatch(state.Out)
	if matches == nil {
		return "", fmt.Errorf("The BMC status in the workflow state is invalid: %s", state.Out)
	}

	return matches[1], nil
}

// changeNodePowerStatus change the power status (on/off) of the node with the given level (soft/hard)
func (d *Driver) changeNodePowerStatus(status string, level string) error {
	node, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	op, err := d.G5kAPI.SubmitPowerOperation(api.PowerOperation{
		Nodes:  []string{node},
		Status: status,
		Level:  level,
	})

	if err != nil {
		return err
	}

	log.Infof("Power-%s (%s) operation for '%s' node have been submitted successfully (workflow id: '%s')", status, level, node, op.WID)
	return d.waitUntilWorkflowIsDone("power", op.WID, node)
}

// rebootNode reboot the node with the given level (soft/hard)
func (d *Driver) rebootNode(level string) error {
	node, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	op, err := d.G5kAPI.SubmitRebootOperation(api.RebootOperation{
		Kind:  "simple",
		Nodes: []string{node},
		Level: level,
	})

	if err != nil {
		return err
	}

	log.Infof("Reboot (%s) operation for '%s' node have been submitted successfully (workflow id: '%s')", level, node, op.WID)
	return d.waitUntilWorkflowIsDone("reboot", op.WID, node)
}
