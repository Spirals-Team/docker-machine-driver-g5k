package driver

import (
	"fmt"
	"net"

	"github.com/docker/machine/libmachine/mcnutils"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	gossh "golang.org/x/crypto/ssh"
)

// g5kReferenceEnvironment is the name of the reference environment automatically deployed on the node by Grid'5000
const g5kReferenceEnvironmentName string = "debian9-x64-std"

// Driver parameters
type Driver struct {
	*drivers.BaseDriver

	G5kAPI                             *api.Client
	G5kJobID                           int
	G5kUsername                        string
	G5kPassword                        string
	G5kSite                            string
	G5kWalltime                        string
	G5kImage                           string
	G5kResourceProperties              string
	G5kSkipVpnChecks                   bool
	G5kReuseRefEnvironment             bool
	G5kJobQueue                        string
	G5kJobStartTime                    string
	DriverSSHPublicKey                 string
	ExternalSSHPublicKeys              []string
	G5kKeepAllocatedResourceAtDeletion bool
}

// NewDriver creates and returns a new instance of the driver
func NewDriver() *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser: drivers.DefaultSSHUser,
			SSHPort: drivers.DefaultSSHPort,
		},
	}
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "g5k"
}

// GetCreateFlags add command line flags to configure the driver
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "G5K_USERNAME",
			Name:   "g5k-username",
			Usage:  "Your Grid5000 account username",
			Value:  "",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_PASSWORD",
			Name:   "g5k-password",
			Usage:  "Your Grid5000 account password",
			Value:  "",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_SITE",
			Name:   "g5k-site",
			Usage:  "Site to reserve the resources on",
			Value:  "",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_WALLTIME",
			Name:   "g5k-walltime",
			Usage:  "Machine's lifetime (HH:MM:SS)",
			Value:  "1:00:00",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_IMAGE",
			Name:   "g5k-image",
			Usage:  "Name of the image (environment) to deploy on the node",
			Value:  g5kReferenceEnvironmentName,
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_RESOURCE_PROPERTIES",
			Name:   "g5k-resource-properties",
			Usage:  "Resource selection with OAR properties (SQL format)",
		},

		mcnflag.BoolFlag{
			EnvVar: "G5K_SKIP_VPN_CHECKS",
			Name:   "g5k-skip-vpn-checks",
			Usage:  "Skip the VPN client connection and DNS configuration checks (for specific use case only, you should not enable this flag in normal use)",
		},

		mcnflag.BoolFlag{
			EnvVar: "G5K_REUSE_REF_ENVIRONMENT",
			Name:   "g5k-reuse-ref-environment",
			Usage:  "Reuse the Grid'5000 reference environment instead of re-deploying the node (it saves a lot of time)",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_JOB_QUEUE",
			Name:   "g5k-job-queue",
			Usage:  "Specify the job queue (besteffort is NOT supported)",
			Value:  "default",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_MAKE_RESOURCE_RESERVATION",
			Name:   "g5k-make-resource-reservation",
			Usage:  "Make a resource reservation for the given start date. (in either 'YYYY-MM-DD HH:MM:SS' date format or an UNIX timestamp)",
		},

		mcnflag.IntFlag{
			EnvVar: "G5K_USE_RESOURCE_RESERVATION",
			Name:   "g5k-use-resource-reservation",
			Usage:  "Use a resource reservation (need to be a job of 'deploy' type and in the 'running' state)",
		},

		mcnflag.StringSliceFlag{
			EnvVar: "G5K_EXTERNAL_SSH_PUBLIC_KEYS",
			Name:   "g5k-external-ssh-public-keys",
			Usage:  "Additional SSH public key(s) allowed to connect to the node (in authorized_keys format)",
			Value:  []string{},
		},

		mcnflag.BoolFlag{
			EnvVar: "G5K_KEEP_RESOURCE_AT_DELETION",
			Name:   "g5k-keep-resource-at-deletion",
			Usage:  "Keep the allocated resource when removing the machine (the job will NOT be killed)",
		},
	}
}

// SetConfigFromFlags configure the driver from the command line arguments
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.G5kUsername = opts.String("g5k-username")
	d.G5kPassword = opts.String("g5k-password")
	d.G5kSite = opts.String("g5k-site")
	d.G5kWalltime = opts.String("g5k-walltime")
	d.G5kImage = opts.String("g5k-image")
	d.G5kResourceProperties = opts.String("g5k-resource-properties")
	d.G5kSkipVpnChecks = opts.Bool("g5k-skip-vpn-checks")
	d.G5kReuseRefEnvironment = opts.Bool("g5k-reuse-ref-environment")
	d.G5kJobQueue = opts.String("g5k-job-queue")
	d.G5kJobStartTime = opts.String("g5k-make-resource-reservation")
	d.G5kJobID = opts.Int("g5k-use-resource-reservation")
	d.ExternalSSHPublicKeys = opts.StringSlice("g5k-external-ssh-public-keys")
	d.G5kKeepAllocatedResourceAtDeletion = opts.Bool("g5k-keep-resource-at-deletion")

	// Docker Swarm
	d.BaseDriver.SetSwarmConfigFromFlags(opts)

	// username is required
	if d.G5kUsername == "" {
		return fmt.Errorf("You must give your Grid5000 account username")
	}

	// password is required
	if d.G5kPassword == "" {
		return fmt.Errorf("You must give your Grid5000 account password")
	}

	// site is required
	if d.G5kSite == "" {
		return fmt.Errorf("You must give the site you want to reserve the resources on")
	}

	// contradictory use of parameters: providing an image to deploy while trying to reuse the reference environment
	if d.G5kReuseRefEnvironment && d.G5kImage != g5kReferenceEnvironmentName {
		return fmt.Errorf("You have to choose between reusing the reference environment or redeploying the node with another image")
	}

	// we cannot reuse the reference environment when the job is of type 'deploy'
	if d.G5kReuseRefEnvironment && (d.G5kJobStartTime != "" || d.G5kJobID != 0) {
		return fmt.Errorf("Reusing the Grid'5000 reference environment on a resource reservation is not supported")
	}

	// warn if user disable VPN check
	if d.G5kSkipVpnChecks {
		log.Warn("VPN client connection and DNS configuration checks are disabled")
	}

	// we cannot use the besteffort queue with docker-machine
	if d.G5kJobQueue == "besteffort" {
		return fmt.Errorf("The besteffort queue is not supported")
	}

	return nil
}

// GetIP returns the ip
func (d *Driver) GetIP() (string, error) {
	return d.BaseDriver.GetIP()
}

// GetMachineName returns the machine name
func (d *Driver) GetMachineName() string {
	return d.BaseDriver.GetMachineName()
}

// GetSSHHostname returns the machine hostname
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHKeyPath returns the ssh private key path
func (d *Driver) GetSSHKeyPath() string {
	return d.BaseDriver.GetSSHKeyPath()
}

// GetSSHPort returns the ssh port
func (d *Driver) GetSSHPort() (int, error) {
	return d.BaseDriver.GetSSHPort()
}

// GetSSHUsername returns the ssh user name
func (d *Driver) GetSSHUsername() string {
	return d.BaseDriver.GetSSHUsername()
}

// GetURL returns the URL of the docker daemon
func (d *Driver) GetURL() (string, error) {
	// get IP address
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	// format URL 'tcp://host:2376'
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

// GetState returns the state of the node
func (d *Driver) GetState() (state.State, error) {
	// get job state from API
	status, err := d.G5kAPI.GetJobState(d.G5kJobID)
	if err != nil {
		return state.Error, err
	}

	switch status {
	case "waiting":
		return state.Starting, nil
	case "launching":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "hold":
		return state.Stopped, nil
	case "error":
		return state.Error, nil
	case "terminated":
		return state.Stopped, nil
	default:
		return state.None, nil
	}
}

// PreCreateCheck check parameters and submit the job to Grid5000
func (d *Driver) PreCreateCheck() error {
	// prepare the driver store dir
	if err := d.prepareDriverStoreDirectory(); err != nil {
		return err
	}

	// check VPN connection if enabled
	if !d.G5kSkipVpnChecks {
		if err := d.checkVpnConnection(); err != nil {
			return err
		}
	}

	// create API client
	d.G5kAPI = api.NewClient(d.G5kUsername, d.G5kPassword, d.G5kSite)

	// load driver SSH public key
	if err := d.loadDriverSSHPublicKey(); err != nil {
		return err
	}

	// check format of external SSH public keys
	for _, externalSSHPubKey := range d.ExternalSSHPublicKeys {
		_, _, _, _, err := gossh.ParseAuthorizedKey([]byte(externalSSHPubKey))
		if err != nil {
			return fmt.Errorf("The external SSH public key '%s' is invalid: %s", externalSSHPubKey, err.Error())
		}
	}

	// skip the job submission/reservation if a job ID is provided
	if d.G5kJobID == 0 {
		if d.G5kJobStartTime == "" {
			// make a job submission: the resources will be reserved for immediate use
			if err := d.makeJobSubmission(); err != nil {
				return err
			}
		} else {
			// make a job reservation: the resources will be reserved for a defined date/time
			if err := d.makeJobReservation(); err != nil {
				return err
			}

			// stop the machine creation
			return fmt.Errorf("The job reservation have been successfully sent. Don't forget to save the Job ID to create the machine when the resources are available")
		}
	}

	// wait for job to be in 'running' state
	if err := d.waitUntilJobIsReady(); err != nil {
		return err
	}

	return nil
}

// Create wait for the job to be running, deploy the OS image and copy the ssh keys
func (d *Driver) Create() error {
	// get node hostname from API
	job, err := d.G5kAPI.GetJob(d.G5kJobID)
	if err != nil {
		return err
	}
	d.BaseDriver.IPAddress = job.Nodes[0]

	// deploy OS image to the node
	if err := d.deployImageToNode(); err != nil {
		return err
	}

	// copy driver SSH key pair to machine directory
	if err := mcnutils.CopyFile(d.getDriverSSHKeyPath(), d.GetSSHKeyPath()); err != nil {
		return err
	}
	if err := mcnutils.CopyFile(d.getDriverSSHKeyPath()+".pub", d.GetSSHKeyPath()+".pub"); err != nil {
		return err
	}

	return nil
}

// Remove delete the resources reservation
func (d *Driver) Remove() error {
	// keep the resource allocated if the user asked for it
	if !d.G5kKeepAllocatedResourceAtDeletion {
		log.Infof("Killing job... (id: '%d')", d.G5kJobID)
		d.G5kAPI.KillJob(d.G5kJobID)
	}

	return nil
}

// Kill don't do anything
func (d *Driver) Kill() error {
	return fmt.Errorf("The 'kill' operation is not supported on Grid'5000")
}

// Start don't do anything
func (d *Driver) Start() error {
	return fmt.Errorf("The 'start' operation is not supported on Grid'5000")
}

// Stop don't do anything
func (d *Driver) Stop() error {
	return fmt.Errorf("The 'stop' operation is not supported on Grid'5000")
}

// Restart don't do anything
func (d *Driver) Restart() error {
	return fmt.Errorf("The 'restart' operation is not supported on Grid'5000")
}
