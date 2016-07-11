package driver

import (
    "errors"
    "fmt"
    "path/filepath"
    "os"
    "docker-machine-driver-g5k/api"

    "github.com/docker/machine/libmachine/drivers"
    "github.com/docker/machine/libmachine/state"
    "github.com/docker/machine/libmachine/mcnflag"
    "github.com/docker/machine/libmachine/mcnutils"
    "github.com/docker/machine/libmachine/log"
)

type Driver struct {
    *drivers.BaseDriver
    *api.Api

    JobId      int
    g5kUser    string
    g5kPasswd  string
    g5kSite    string
}

func NewDriver() *Driver {
    return &Driver{
        BaseDriver: &drivers.BaseDriver{
            SSHPort: drivers.DefaultSSHPort,
        },
    }
}

// TODO To complete
func (d *Driver) Create() (err error) {
    var job *api.Job

    client := d.getApi()
    if job, err = client.GetJob(d.JobId); err != nil {
        return err
    }

    sshport, _ := d.GetSSHPort()
    d.BaseDriver.IPAddress = job.Nodes[0]
    d.BaseDriver.SSHArgs = []string{"-o", fmt.Sprintf("ProxyCommand ssh %s@access.grid5000.fr -W %s:%v", d.g5kUser, d.BaseDriver.IPAddress, sshport)}

    home := mcnutils.GetHomeDir()
    src, dst := filepath.Join(home, ".ssh/id_rsa"), d.GetSSHKeyPath()

    if err = mcnutils.CopyFile(src, dst); err != nil {
        return err
    }
    if err = os.Chmod(dst, 0600); err != nil {
        return err
    }

    log.Debug(d.BaseDriver)

    return nil
}

func (d *Driver) DriverName() string {
    return "g5k"
}

func (d *Driver) getApi() *api.Api {
    if d.Api == nil {
        d.Api = api.NewApi(d.g5kUser, d.g5kPasswd, d.g5kSite)
    }
    return d.Api
}

// TODO To complete
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
    return []mcnflag.Flag{
        mcnflag.StringFlag{
            Name:   "g5k-username",
            Usage:  "Username account to log on G5K grid",
            Value:  "",
        },
        mcnflag.StringFlag{
            Name:   "g5k-passwd",
            Usage:  "G5K user's account's password",
            Value:  "",
        },
        mcnflag.StringFlag{
            Name:   "g5k-site",
            Usage:  "Name of the site to connect to",
            Value:  "",
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
        case "waiting"    :
            return state.Starting, nil
        case "launching"  :
            return state.Starting, nil
        case "running"    :
            return state.Running, nil
        case "hold"       :
            return state.Stopped, nil
        case "error"      :
            return state.Error, nil
        case "terminated" :
            return state.Stopped, nil
        default           :
            return state.None, nil
    }
}

// TODO To implement
func (d *Driver) Kill() error {
    return nil
}

// TODO To complete
func (d *Driver) PreCreateCheck() (err error) {
    if d.g5kUser == "" {
        return errors.New("You must give your g5k account")
    }
    if d.g5kPasswd == "" {
        return errors.New("You must give your g5k password")
    }
    if d.g5kSite == "" {
        return errors.New("You must give the site you want to log on")
    }

    client := d.getApi()

    log.Info("Submitting job...")
    if d.JobId, err = client.SubmitJob(); err != nil {
        return err
    }
    log.Info("Nodes allocated and ready")

    log.Info("Deploying environment")
    if err = client.DeployEnvironment(d.JobId); err != nil {
        return err
    }
    log.Info("Environment deployed")

    return nil
}

// TODO To implement
func (d *Driver) Remove() error {
    client := d.getApi()
    return client.KillJob(d.JobId)
}

// TODO To implement
func (d *Driver) Restart() error {
    return nil
}

// TODO To complete
func (d *Driver) SetConfigFromFlags(opts drivers.DriverOptions) error {
    d.g5kUser = opts.String("g5k-username")
    d.g5kPasswd = opts.String("g5k-passwd")
    d.g5kSite = opts.String("g5k-site")

    // We log on the node as root
    d.BaseDriver.SSHUser = "root"

    // Docker Swarm
    d.BaseDriver.SetSwarmConfigFromFlags(opts)
    return nil
}

// TODO To implement
func (d *Driver) Start() error {
    return nil
}

// TODO To implement
func (d *Driver) Stop() error {
    return nil
}
