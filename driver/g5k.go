package driver

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"time"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/docker/machine/libmachine/log"
)

func (d *Driver) checkVpnConfiguration() error {
	// Check VPN connection by trying to connect to the ssh server of the frontend of the current site.
	// This allows to test if the user use the VPN and the Grid'5000 DNS servers.
	if err := CheckSSHConnection(fmt.Sprintf("frontend.%s.grid5000.fr", d.G5kSite)); err != nil {
		return fmt.Errorf("Connection to frontend of '%s' site failed. Please check if the site is not undergoing maintenance and your VPN client is connected and properly configured (see driver documentation for more information)", d.G5kSite)
	}

	return nil
}

// waitUntilJobIsReady wait until the job reach the 'running' state (no timeout)
func (d *Driver) waitUntilJobIsReady() error {
	log.Info("Waiting for job to run...")

	for {
		// get job
		job, err := d.g5kAPI.GetJob(d.G5kJobID)
		if err != nil {
			return err
		}

		// check if the job is running
		if job.State == "running" {
			break
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
		// convert the ssh authorized_keys to be added in base64
		sshAuthorizedKeysBase64 := base64.StdEncoding.EncodeToString([]byte(GenerateSSHAuthorizedKeys(d.DriverSSHPublicKey, d.ExternalSSHPublicKeys)))
		// enable sudo for current user, add public key to ssh authorized keys for root user and wait the end of the job
		jobCommand = fmt.Sprint(`sudo-g5k && printf ` + sshAuthorizedKeysBase64 + ` |base64 -d |sudo tee -a /root/.ssh/authorized_keys >/dev/null && sleep 365d`)
	}

	// submit new Job request
	jobID, err := d.g5kAPI.SubmitJob(api.JobRequest{
		Resources:  fmt.Sprintf("nodes=1,walltime=%s", d.G5kWalltime),
		Command:    jobCommand,
		Properties: d.G5kResourceProperties,
		Types:      jobTypes,
		Queue:      d.G5kJobQueue,
	})
	if err != nil {
		return fmt.Errorf("Error when submitting new job: %s", err.Error())
	}

	log.Infof("Job submission have been successfully submitted. (job id: %d)", jobID)
	d.G5kJobID = jobID
	return nil
}

// makeJobReservation submit a job reservation to Grid'5000
func (d *Driver) makeJobReservation() error {
	jobCommand := "sleep 365d"
	jobTypes := []string{"deploy"}

	// submit new Job request
	jobID, err := d.g5kAPI.SubmitJob(api.JobRequest{
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

	log.Infof("Job reservation have been successfully submitted. (job id: %d)", jobID)
	d.G5kJobID = jobID
	return nil
}

// waitUntilWorkflowIsDone will wait until the workflow for the given operation is done (successfully or not) for the node
func (d *Driver) waitUntilWorkflowIsDone(operation string, wid string, node string) error {
	log.Infof("Waiting for workflow of '%s' operation to finish, it will take a few minutes...", operation)

	for {
		// get operation workflow
		workflow, err := d.g5kAPI.GetOperationWorkflow(operation, wid)
		if err != nil {
			return err
		}

		// check if the workflow is done for the node
		if ArrayContainsString(workflow.Nodes["ok"], node) {
			break
		}

		// check if the workflow failed for the node
		if ArrayContainsString(workflow.Nodes["ko"], node) {
			return fmt.Errorf("Workflow for '%s' operation failed for the '%s' node", operation, node)
		}

		// check if the workflow is processing the node
		if ArrayContainsString(workflow.Nodes["processing"], node) {
			log.Debugf("Workflow for '%s' operation is in processing state for the '%s' node", operation, node)
		}

		// wait before making another API call
		time.Sleep(7 * time.Second)
	}

	log.Infof("Workflow for '%s' operation finished successfully for the '%s' node", operation, node)
	return nil
}

// deployImageToNode start the deployment of an OS image to a node
func (d *Driver) deployImageToNode() error {
	// if the user want to reuse Grid'5000 reference environment
	if d.G5kReuseRefEnvironment {
		log.Infof("Skipping image deployment and reusing Grid'5000 standard environment")
		return nil
	}

	// get job informations
	job, err := d.g5kAPI.GetJob(d.G5kJobID)
	if err != nil {
		return fmt.Errorf("Error when getting job (id: '%d') informations: %s", d.G5kJobID, err.Error())
	}

	// check job type before deploying
	if !ArrayContainsString(job.Types, "deploy") {
		return fmt.Errorf("The job (id: %d) needs to have the type 'deploy'", d.G5kJobID)
	}

	// get the hostname of the node
	node, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	// check if the node is allocated to the job
	if !ArrayContainsString(job.Nodes, node) {
		return fmt.Errorf("The node '%s' is not allocated to the job (id: %d)", node, d.G5kJobID)
	}

	log.Infof("Submitting a new deployment for node '%s'... (image: '%s')", node, d.G5kImage)

	// convert the ssh authorized_keys to be added in base64
	sshAuthorizedKeysBase64 := base64.StdEncoding.EncodeToString([]byte(GenerateSSHAuthorizedKeys(d.DriverSSHPublicKey, d.ExternalSSHPublicKeys)))

	// submit deployment operation to kadeploy
	op, err := d.g5kAPI.SubmitDeployment(api.DeploymentOperation{
		Nodes: []string{node},
		Environment: api.DeploymentOperationEnvironment{
			Kind: "database",
			Name: d.G5kImage,
		},
		CustomOperations: map[string]map[string]map[string][]api.DeploymentOperationCustomOperation{
			"BroadcastEnvKascade": {
				"manage_user_post_install": {
					"post-ops": {
						api.DeploymentOperationCustomOperation{
							Name:    "docker_machine_driver_ssh_root_pub_keys",
							Action:  "exec",
							Command: fmt.Sprint(`printf ` + sshAuthorizedKeysBase64 + ` |base64 -d >> $KADEPLOY_ENV_EXTRACTION_DIR/root/.ssh/authorized_keys`),
						},
					},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("Error when submitting new deployment: %s", err.Error())
	}

	log.Infof("Deployment operation for '%s' node have been submitted successfully (workflow id: '%s')", node, op.WID)

	// waiting deployment to finish (REQUIRED or you will interfere with kadeploy)
	if err = d.waitUntilWorkflowIsDone("deployment", op.WID, node); err != nil {
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

	op, err := d.g5kAPI.RequestPowerStatus(node)
	if err != nil {
		return "", fmt.Errorf("Failed to request power status: %s", err.Error())
	}

	if err := d.waitUntilWorkflowIsDone("power", op.WID, node); err != nil {
		return "", err
	}

	// get nodes states for the workflow
	states, err := d.g5kAPI.GetOperationStates("power", op.WID)
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
	if d.G5kReuseRefEnvironment {
		return fmt.Errorf("You can't power-%s (%s) the node when reusing the Grid'5000 environment", status, level)
	}

	node, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	op, err := d.g5kAPI.SubmitPowerOperation(api.PowerOperation{
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
	if d.G5kReuseRefEnvironment {
		return fmt.Errorf("You can't reboot (%s) the node when reusing the Grid'5000 environment", level)
	}

	node, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("Failed to get the node hostname: %s", err.Error())
	}

	op, err := d.g5kAPI.SubmitRebootOperation(api.RebootOperation{
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
