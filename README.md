# docker-machine-driver-g5k
A Docker Machine driver for the Grid5000 testbed infrastructure. It will create  a Docker machine on a node of the Grid5000.

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine/)
* [Go tools](https://golang.org/doc/install)

You will also need a Grid5000 account to use the driver. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation
*The procedure was only tested on Ubuntu 16.04.*

### Local host
To use the Go tools, you need to set your [GOPATH](https://golang.org/doc/code.html#GOPATH) variable environment.

To compile the binary, run:

```bash
mkdir $GOPATH/src
mv the/sources/dir/docker-machine-driver-g5k $GOPATH/src
go get -d -u github.com/docker/machine
cd $GOPATH/src/docker-machine-driver-g5k
go install
```

Then, either put the driver in a directory filled in your PATH environment variable, or run:

```bash
export PATH=/my/driver/dir/:$PATH
```

Be sure that your SSH public keys file is `$HOME/.ssh/id_rsa.pub`

### Grid5000
You should have an `authorized_keys` file in your SSH directory. Put it on your public directory, on each site you wish to create a machine on.

## How to use
### Options
The driver will need a few options to create a machine. Here is a list of options:

|       Option      |  Description     |  Default value   | Required  |
|-------------------|------------------|------------------|-----------|
| `--g5k-username`  | User's account   |                  | Yes       |
| `--g5k-passwd`    | User's passwd    |                  | Yes       |
| `--g5k-site`      | Site's location of the machine |    | Yes       |
| `--g5k-walltime`  | Timelife of the machine | "1:00:00" | No        |

### Example
An example :

```bash
docker-machine create -d g5k \
--g5k-username user \
--g5k-passwd ******** \
--g5k-site lille \
--g5k-walltime 2:45:00
```

At the end, this error should happen:

`Error creating machine: Error checking the host: Error checking and/or regenerating the certs: There was an error validating certificates for host "node.site.grid5000.fr:2376": dial tcp: lookup node.site.grid5000.fr on 127.0.1.1:53: no such host`

You can ignore it: the machine is ready and accessible.

### VPN
To connect to your Docker machine, you will (probably) need to connect to the Grid5000 VPN. More details [here](https://www.grid5000.fr/mediawiki/index.php/VPN)
