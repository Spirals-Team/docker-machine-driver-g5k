# docker-machine-driver-g5k
A Docker Machine driver for the Grid5000 testbed infrastructure. It will provision a Docker machine on a node of the Grid5000.

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine/)
* [Go tools](https://golang.org/doc/install)

You need a Grid5000 account to use this driver. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation
*This procedure was tested on Ubuntu 16.04 and MacOS.*

To use the Go tools, you need to set your [GOPATH](https://golang.org/doc/code.html#GOPATH) variable environment.

To get the code and compile the binary, run:

```bash
go get -u github.com/Spirals-Team/docker-machine-driver-g5k
```

Then, either put the driver in a directory filled in your PATH environment variable, or run:

```bash
export PATH=$PATH:$GOPATH/bin
```

## How to use

### VPN
You need to be connected to the Grid5000 VPN to create and access your Docker node.  
Do not forget to configure your DNS or use OpenVPN DNS auto-configuration.  
Please follow the instructions on the [Grid5000 Wiki](https://www.grid5000.fr/mediawiki/index.php/VPN).

### Options
The driver needs a few options to create a machine. Here is a list of options:

|          Option          |              Description              |     Default value     |  Required  |
|--------------------------|---------------------------------------|-----------------------|------------|
| `--g5k-username`         | Your Grid5000 account username        |                       | Yes        |
| `--g5k-password`         | Your Grid5000 account password        |                       | Yes        |
| `--g5k-site`             | Site to reserve the resources on      |                       | Yes        |
| `--g5k-walltime`         | Timelife of the machine               | "1:00:00"             | No         |
| `--g5k-ssh-private-key`  | Path of your ssh private key          | "~/.ssh/id_rsa"       | No         |
| `--g5k-ssh-public-key`   | Path of your ssh public key           | "< private-key >.pub" | No         |
| `--g5k-image`            | Name of the image to deploy           | "jessie-x64-min"      | No         |

### Example
An example of node provisioning :

```bash
docker-machine create -d g5k \
--g5k-username user \
--g5k-password ******** \
--g5k-site lille \
--g5k-walltime 2:45:00 \
--g5k-ssh-private-key ~/.ssh/g5k-key \
test-node
```
