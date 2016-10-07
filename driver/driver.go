package driver

import (
	"errors"
	"fmt"
	"os"

	"github.com/Spirals-Team/docker-machine-driver-g5k/api"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	*api.Api

	JobID                 int
	G5kUsername           string
	G5kPassword           string
	G5kSite               string
	g5kWalltime           string
	g5kSSHPrivateKeyPath  string
	g5kSSHPublicKeyPath   string
	g5kImage              string
	g5kResourceProperties string
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

func (d *Driver) getAPI() *api.Api {
	if d.Api == nil {
		d.Api = api.NewApi(d.G5kUsername, d.G5kPassword, d.G5kSite, d.g5kImage)
	}
	return d.Api
}

// GetCreateFlags add command line flags to configure the driver
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "g5k-username",
			Usage: "Your Grid5000 account username",
			Value: "",
		},

		mcnflag.StringFlag{
			Name:  "g5k-password",
			Usage: "Your Grid5000 account password",
			Value: "",
		},

		mcnflag.StringFlag{
			Name:  "g5k-site",
			Usage: "Site to reserve the resources on",
			Value: "",
		},

		mcnflag.StringFlag{
			Name:  "g5k-walltime",
			Usage: "Machine's lifetime (HH:MM:SS)",
			Value: "1:00:00",
		},

		mcnflag.StringFlag{
			Name:  "g5k-ssh-private-key",
			Usage: "Path of your ssh private key",
			Value: mcnutils.GetHomeDir() + "/.ssh/id_rsa",
		},

		mcnflag.StringFlag{
			Name:  "g5k-ssh-public-key",
			Usage: "Path of your ssh public key",
			Value: "",
		},

		mcnflag.StringFlag{
			Name:  "g5k-image",
			Usage: "Name of the image to deploy",
			Value: "jessie-x64-min",
		},

		mcnflag.StringFlag{
			Name:  "g5k-resource-properties",
			Usage: "Resource selection with OAR properties (SQL format)",
			Value: "",
		},
	}
}

// SetConfigFromFlags configure the driver from the command line arguments
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.G5kUsername = opts.String("g5k-username")
	d.G5kPassword = opts.String("g5k-password")
	d.G5kSite = opts.String("g5k-site")
	d.g5kWalltime = opts.String("g5k-walltime")
	d.g5kSSHPrivateKeyPath = opts.String("g5k-ssh-private-key")

	// if the user dont specify a public key path, append .pub to the private key path
	if opts.String("g5k-ssh-public-key") != "" {
		d.g5kSSHPublicKeyPath = opts.String("g5k-ssh-public-key")
	} else {
		d.g5kSSHPublicKeyPath = d.g5kSSHPrivateKeyPath + ".pub"
	}

	d.g5kImage = opts.String("g5k-image")
	d.g5kResourceProperties = opts.String("g5k-resource-properties")

	// Docker Swarm
	d.BaseDriver.SetSwarmConfigFromFlags(opts)

	return nil
}

// GetIP returns the ip
func (d *Driver) GetIP() (string, error) {
	return d.BaseDriver.IPAddress, nil
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
	url, err := d.BaseDriver.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s:2376", url), nil
}

// GetState returns the state of the node
func (d *Driver) GetState() (state.State, error) {
	client := d.getAPI()

	status, err := client.GetJobState(d.JobID)
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
	if d.G5kUsername == "" {
		return errors.New("You must give your Grid5000 account username")
	}
	if d.G5kPassword == "" {
		return errors.New("You must give your Grid5000 account password")
	}
	if d.G5kSite == "" {
		return errors.New("You must give the site you want to reserve the resources on")
	}
	if _, err := os.Stat(d.g5kSSHPrivateKeyPath); os.IsNotExist(err) {
		return errors.New("Your ssh private key file does not exist in : '" + d.g5kSSHPrivateKeyPath + "'")
	}
	if _, err := os.Stat(d.g5kSSHPublicKeyPath); os.IsNotExist(err) {
		return errors.New("Your ssh public key file does not exist in : ''" + d.g5kSSHPublicKeyPath + "'")
	}

	client := d.getAPI()

	log.Info("Submitting job...")
	if d.JobID, err = client.SubmitJob(d.g5kWalltime, d.g5kResourceProperties); err != nil {
		return err
	}
	log.Info("Nodes allocated and ready")

	log.Info("Deploying environment. It will take a few minutes...")
	if err = client.DeployEnvironment(d.JobID, d.g5kSSHPublicKeyPath); err != nil {
		return err
	}
	log.Info("Environment deployed")

	return nil
}

// Create deploy the environment and create the Docker machine
func (d *Driver) Create() (err error) {
	// Get IP address from API
	client := d.getAPI()
	if job, err := client.GetJob(d.JobID); err != nil {
		return err
	} else {
		d.BaseDriver.IPAddress = job.Nodes[0]
	}

	// Copy the SSH private key to the docker machine config folder
	src, dst := d.g5kSSHPrivateKeyPath, d.GetSSHKeyPath()
	if err = mcnutils.CopyFile(src, dst); err != nil {
		return err
	}

	return nil
}

// Remove delete the resources reservation
func (d *Driver) Remove() error {
	client := d.getAPI()
	log.Info("Killing job...")
	client.KillJob(d.JobID)

	// We get an error if the job was already dead, which is not really an error
	return nil
}

// TODO To implement
func (d *Driver) Kill() error {
	return fmt.Errorf("Cannot kill a machine on G5K")
}

func (d *Driver) Start() error {
	return fmt.Errorf("Cannot start a machine on G5K")
}

func (d *Driver) Stop() error {
	return fmt.Errorf("Cannot stop a machine on G5K")
}

func (d *Driver) Restart() error {
	return fmt.Errorf("Cannot restart a machine on G5K")
}
