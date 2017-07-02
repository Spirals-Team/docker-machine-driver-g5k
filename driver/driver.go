package driver

import (
	"fmt"
	"net"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

// Driver parameters
type Driver struct {
	*drivers.BaseDriver

	G5kAPI                *api.Client
	G5kJobID              int
	G5kUsername           string
	G5kPassword           string
	G5kSite               string
	G5kWalltime           string
	G5kImage              string
	G5kResourceProperties string
	G5kHostToProvision    string
	G5kSkipVpnChecks      bool
	SSHKeyPair            *ssh.KeyPair
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
			Usage:  "Name of the image to deploy",
			Value:  "jessie-x64-min",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_RESOURCE_PROPERTIES",
			Name:   "g5k-resource-properties",
			Usage:  "Resource selection with OAR properties (SQL format)",
			Value:  "",
		},

		mcnflag.IntFlag{
			EnvVar: "G5K_USE_JOB_RESERVATION",
			Name:   "g5k-use-job-reservation",
			Usage:  "job ID to use (need to be an already existing job ID, because job reservation will be skipped)",
		},

		mcnflag.StringFlag{
			EnvVar: "G5K_HOST_TO_PROVISION",
			Name:   "g5k-host-to-provision",
			Usage:  "Host to provision (host need to be already deployed, because deployment step will be skipped)",
			Value:  "",
		},

		mcnflag.BoolFlag{
			EnvVar: "G5K_SKIP_VPN_CHECKS",
			Name:   "g5k-skip-vpn-checks",
			Usage:  "Skip the VPN client connection and DNS configuration checks (for particular use case only, you should not enable this flag in normal use)",
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
	d.G5kJobID = opts.Int("g5k-use-job-reservation")
	d.G5kHostToProvision = opts.String("g5k-host-to-provision")
	d.G5kSkipVpnChecks = opts.Bool("g5k-skip-vpn-checks")

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

	// warn if user disable VPN check
	if d.G5kSkipVpnChecks {
		log.Warn("VPN client connection and DNS configuration checks are disabled")
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
func (d *Driver) PreCreateCheck() (err error) {
	// check VPN connection if enabled
	if err := d.checkVpnConnection(); !d.G5kSkipVpnChecks && (err != nil) {
		return err
	}

	// create API client
	d.G5kAPI = api.NewClient(d.G5kUsername, d.G5kPassword, d.G5kSite)

	// submit new job reservation
	if err := d.submitNewJobReservation(); err != nil {
		return err
	}

	// check if a SSH key pair is available
	if d.SSHKeyPair == nil {
		// generate a new SSH key pair
		d.SSHKeyPair, err = ssh.NewKeyPair()
		if err != nil {
			return fmt.Errorf("Error when generating a new SSH key pair: %s", err.Error())
		}
	}

	// submit new deployment
	if err := d.submitNewDeployment(); err != nil {
		return err
	}

	return nil
}

// Create copy ssh key in docker-machine dir and set the node IP
func (d *Driver) Create() (err error) {
	// provisionning only mode
	if d.G5kHostToProvision != "" {
		// use provided hostname
		d.BaseDriver.IPAddress = d.G5kHostToProvision
	} else {
		// get hostname from API
		job, err := d.G5kAPI.GetJob(d.G5kJobID)
		if err != nil {
			return err
		}
		d.BaseDriver.IPAddress = job.Nodes[0]
	}

	// copy SSH key pair to machine directory
	if err := d.SSHKeyPair.WriteToFile(d.GetSSHKeyPath(), d.GetSSHKeyPath()+".pub"); err != nil {
		return fmt.Errorf("Error when copying SSH key pair to machine directory: %s", err.Error())
	}

	return nil
}

// Remove delete the resources reservation
func (d *Driver) Remove() error {
	log.Infof("Killing job... (id: '%d')", d.G5kJobID)

	// send kill job command to API
	d.G5kAPI.KillJob(d.G5kJobID)

	return nil
}

// Kill don't do anything
func (d *Driver) Kill() error {
	return fmt.Errorf("You can't kill a machine on Grid'5000")
}

// Start don't do anything
func (d *Driver) Start() error {
	return fmt.Errorf("You can't start a machine on Grid'5000")
}

// Stop don't do anything
func (d *Driver) Stop() error {
	return fmt.Errorf("You can't stop a machine on Grid'5000")
}

// Restart don't do anything
func (d *Driver) Restart() error {
	return fmt.Errorf("You can't restart a machine on Grid'5000")
}
