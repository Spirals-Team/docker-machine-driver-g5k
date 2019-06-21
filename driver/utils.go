package driver

import (
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ArrayContainsString check if the given string array contains the given string
func ArrayContainsString(array []string, str string) bool {
	for _, v := range array {
		if v == str {
			return true
		}
	}
	return false
}

// GenerateSSHAuthorizedKeys generate the SSH AuthorizedKeys composed of the driver and external user defined key(s)
func GenerateSSHAuthorizedKeys(driverKey string, externalKeys []string) string {
	var authorizedKeysEntries []string

	// add driver key
	authorizedKeysEntries = append(authorizedKeysEntries, "# docker-machine driver g5k - driver key")
	authorizedKeysEntries = append(authorizedKeysEntries, driverKey)

	// add external key(s)
	for index, externalPubKey := range externalKeys {
		authorizedKeysEntries = append(authorizedKeysEntries, fmt.Sprintf("# docker-machine driver g5k - additional key %d", index))
		authorizedKeysEntries = append(authorizedKeysEntries, strings.TrimSpace(externalPubKey))
	}

	return strings.Join(authorizedKeysEntries, "\n") + "\n"
}

// CheckSSHConnection will try a SSH connection to the given hostname
func CheckSSHConnection(hostname string) error {
	_, err := ssh.Dial("tcp", net.JoinHostPort(hostname, "22"), &ssh.ClientConfig{Timeout: time.Second * 2})

	// we need to check if the error is network-related because the SSH Dial will always return an error due to the Authentication being not configured
	if _, ok := err.(*net.OpError); ok {
		return fmt.Errorf("Failed to connect to the SSH server on the node '%s' using port 22", hostname)
	}

	// ignore other errors because the ssh Dial will always return an error as there is no auth method configured
	return nil
}
