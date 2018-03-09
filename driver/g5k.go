package driver

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"
	"github.com/docker/machine/libmachine/log"
)

// CheckVpnConnection check if the VPN is connected and properly configured (DNS) by trying to connect to the site frontend SSH server using its hostname
func CheckVpnConnection(site string) error {
	// construct site frontend hostname
	frontend := fmt.Sprintf("frontend.%s.grid5000.fr:22", site)

	// try to connect to the frontend SSH server
	sshConfig := &ssh.ClientConfig{}
	_, err := ssh.Dial("tcp", frontend, sshConfig)

	// we need to check if the error is network-related because the SSH Dial will always return an error due to the Authentication being not configured
	if _, ok := err.(*net.OpError); ok {
		return fmt.Errorf("Connection to frontend of '%s' site failed. Please check if the site is not undergoing maintenance and your VPN client is connected and properly configured (see driver documentation for more information)", site)
	}

	return nil
}

func (d *Driver) submitNewJobReservation() error {
	// if a job ID is provided, skip job reservation
	if d.G5kJobID != 0 {
		log.Infof("Skipping job reservation and using job ID '%v'", d.G5kJobID)
		return nil
	}

	// by default, the node will be redeployed with another image, no specific actions are needed
	jobCommand := "sleep 365d"
	jobTypes := []string{"deploy"}

	// if the user want to reuse the reference environment, specific actions are needed
	if d.G5kReuseRefEnvironment {
		// remove the 'deploy' job type because we will not deploy the machine
		jobTypes = []string{}
		// enable sudo for current user, add public key to ssh authorized keys for root user and wait the end of the job
		jobCommand = `sudo-g5k && sudo /bin/sh -c 'echo "` + string(d.SSHKeyPair.PublicKey) + `" >>/root/.ssh/authorized_keys' && sleep 365d`
	}

	// submit new Job request
	jobID, err := d.G5kAPI.SubmitJob(api.JobRequest{
		Resources:  fmt.Sprintf("nodes=1,walltime=%s", d.G5kWalltime),
		Command:    jobCommand,
		Properties: d.G5kResourceProperties,
		Types:      jobTypes,
	})
	if err != nil {
		return fmt.Errorf("Error when submitting new job: %s", err.Error())
	}

	if err = d.G5kAPI.WaitUntilJobIsReady(jobID); err != nil {
		return fmt.Errorf("Error when waiting for job to be running: %s", err.Error())
	}

	// job is running, keep its ID for future usage
	d.G5kJobID = jobID
	return nil
}

func (d *Driver) handleDeploymentError() {
	// if deployment fail, we can't recover from this error, so we kill the job
	log.Infof("Unrecoverable error in deployment, killing job ID '%d'...", d.G5kJobID)
	d.G5kAPI.KillJob(d.G5kJobID)
}

func (d *Driver) submitNewDeployment() error {
	// if a host to provision is set, skip host deployment
	if d.G5kHostToProvision != "" {
		log.Infof("Skipping host deployment and provisionning host '%s' only", d.G5kHostToProvision)
		return nil
	}

	// if the use want to reuse Grid'5000 reference environment
	if d.G5kReuseRefEnvironment {
		log.Infof("Skipping host deployment and reusing Grid'5000 standard environment")
		return nil
	}

	// get job informations
	job, err := d.G5kAPI.GetJob(d.G5kJobID)
	if err != nil {
		return fmt.Errorf("Error when getting job (id: '%d') informations: %s", d.G5kJobID, err.Error())
	}

	// deploy environment
	deploymentID, err := d.G5kAPI.SubmitDeployment(api.DeploymentRequest{
		Nodes:       job.Nodes,
		Environment: d.G5kImage,
		Key:         string(d.SSHKeyPair.PublicKey),
	})
	if err != nil {
		d.handleDeploymentError()
		return fmt.Errorf("Error when submitting new deployment: %s", err.Error())
	}

	// waiting deployment to finish (REQUIRED or you will interfere with kadeploy)
	if err = d.G5kAPI.WaitUntilDeploymentIsFinished(deploymentID); err != nil {
		d.handleDeploymentError()
		return fmt.Errorf("Error when waiting for deployment to finish: %s", err.Error())
	}

	return nil
}
