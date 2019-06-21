package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/machine/libmachine/ssh"
)

// resolveDriverStorePath returns the store path of the driver
func (d *Driver) resolveDriverStorePath(file string) string {
	return filepath.Join(d.StorePath, "g5k", file)
}

// prepareDriverStoreDirectory initialize the driver storage directory
func (d *Driver) prepareDriverStoreDirectory() error {
	driverStoreBasePath := d.resolveDriverStorePath(".")

	// create the directory if needed
	if _, err := os.Stat(driverStoreBasePath); os.IsNotExist(err) {
		if err := os.Mkdir(driverStoreBasePath, 0700); err != nil {
			return fmt.Errorf("Failed to create the driver storage directory: %s", err)
		}
	}

	return nil
}

// getDriverSSHKeyPath returns the path leading to the driver SSH private key (append .pub to get the public key)
func (d *Driver) getDriverSSHKeyPath() string {
	return d.resolveDriverStorePath("id_rsa")
}

// loadDriverSSHPublicKey load the driver SSH Public key from the storage dir, the key will be created if needed
func (d *Driver) loadDriverSSHPublicKey() error {
	driverSSHKeyPath := d.getDriverSSHKeyPath()

	// generate the driver SSH key pair if needed
	if _, err := os.Stat(driverSSHKeyPath); os.IsNotExist(err) {
		if err := ssh.GenerateSSHKey(driverSSHKeyPath); err != nil {
			return fmt.Errorf("Failed to generate the driver ssh key: %s", err)
		}
	}

	// load the public key from file
	sshPublicKey, err := ioutil.ReadFile(d.getDriverSSHKeyPath() + ".pub")
	if err != nil {
		return fmt.Errorf("Failed to load the driver ssh public key: %s", err)
	}

	// store the public key for future use
	d.DriverSSHPublicKey = strings.TrimSpace(string(sshPublicKey))
	return nil
}
