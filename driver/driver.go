package driver

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

	JobId       int
	G5kUser     string
	G5kPasswd   string
	G5kSite     string
	g5kWalltime string
}

func NewDriver() *Driver {
	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHPort: drivers.DefaultSSHPort,
		},
	}
}

// Achieve the last settings
func (d *Driver) Create() (err error) {
	var job *api.Job

	client := d.getApi()
	if job, err = client.GetJob(d.JobId); err != nil {
		return err
	}

	sshport, _ := d.GetSSHPort()
	d.BaseDriver.IPAddress = job.Nodes[0]
	d.BaseDriver.SSHArgs = []string{"-o", fmt.Sprintf("ProxyCommand ssh %s@access.grid5000.fr -W %s:%v", d.G5kUser, d.BaseDriver.IPAddress, sshport)}

	// Copy the user's SSH private key to the machine folder
	home := mcnutils.GetHomeDir()
	src, dst := filepath.Join(home, ".ssh/id_rsa"), d.GetSSHKeyPath()

	if err = mcnutils.CopyFile(src, dst); err != nil {
		return err
	}
	if err = os.Chmod(dst, 0600); err != nil {
		return err
	}

	return nil
}

func (d *Driver) DriverName() string {
	return "g5k"
}

func (d *Driver) getApi() *api.Api {
	if d.Api == nil {
		d.Api = api.NewApi(d.G5kUser, d.G5kPasswd, d.G5kSite)
	}
	return d.Api
}

// TODO To complete
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:  "g5k-username",
			Usage: "Username account to log on G5K grid",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "g5k-passwd",
			Usage: "G5K user's account's password",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "g5k-site",
			Usage: "Name of the site to connect to",
			Value: "",
		},
		mcnflag.StringFlag{
			Name:  "g5k-walltime",
			Usage: "Machine's lifetime (HH:MM:SS)",
			Value: "1:00:00",
		},
	}
}

func (d *Driver) GetIP() (string, error) {
	return d.BaseDriver.IPAddress, nil
}

func (d *Driver) GetMachineName() string {
	return d.BaseDriver.GetMachineName()
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return d.BaseDriver.GetSSHKeyPath()
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.BaseDriver.GetSSHPort()
}

func (d *Driver) GetSSHUsername() string {
	return d.BaseDriver.GetSSHUsername()
}

func (d *Driver) GetURL() (string, error) {
	url, err := d.BaseDriver.GetIP()

	if err != nil {
		return "", err
	} else {
		url = fmt.Sprintf("tcp://%s:2376", url)
	}

	return url, nil
}

func (d *Driver) GetState() (state.State, error) {
	client := d.getApi()

	status, err := client.GetJobState(d.JobId)
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

// TODO To implement
func (d *Driver) Kill() error {
	return fmt.Errorf("Cannot kill a machine on G5K")
}

// Submit a job and deploy an environment on G5K
func (d *Driver) PreCreateCheck() (err error) {
	if d.G5kUser == "" {
		return errors.New("You must give your G5K account")
	}
	if d.G5kPasswd == "" {
		return errors.New("You must give your G5K password")
	}
	if d.G5kSite == "" {
		return errors.New("You must give the site you want to log on")
	}

	client := d.getApi()

	log.Info("Submitting job...")
	if d.JobId, err = client.SubmitJob(d.g5kWalltime); err != nil {
		return err
	}
	log.Info("Nodes allocated and ready")

	log.Info("Deploying environment. It will take a few minutes...")
	if err = client.DeployEnvironment(d.JobId); err != nil {
		return err
	}
	log.Info("Environment deployed")

	return nil
}

func (d *Driver) Remove() error {
	client := d.getApi()
	log.Info("Killing job...")
	client.KillJob(d.JobId)

	// We get an error if the job was already dead, which is not really an error
	return nil
}

func (d *Driver) Restart() error {
	return fmt.Errorf("Cannot restart a machine on G5K")
}

// TODO To complete
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
	d.G5kUser = opts.String("g5k-username")
	d.G5kPasswd = opts.String("g5k-passwd")
	d.G5kSite = opts.String("g5k-site")
	d.g5kWalltime = opts.String("g5k-walltime")

	// We log on the node as root
	d.BaseDriver.SSHUser = "root"

	// Docker Swarm
	d.BaseDriver.SetSwarmConfigFromFlags(opts)
	return nil
}

func (d *Driver) Start() error {
	return fmt.Errorf("Cannot start a machine on G5K")
}

func (d *Driver) Stop() error {
	return fmt.Errorf("Cannot stop a machine on G5K")
}
