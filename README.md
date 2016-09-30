# docker-machine-driver-g5k
A Docker Machine driver for the Grid5000 testbed infrastructure. It will create a Docker machine on a node of the Grid5000.

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine/)
* [Go tools](https://golang.org/doc/install)

You need a Grid5000 account to use this driver. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation
*The procedure was only tested on Ubuntu 16.04.*

### Local host
To use the Go tools, you need to set your [GOPATH](https://golang.org/doc/code.html#GOPATH) variable environment.

To get the code and compile the binary, run:

```bash
go get -u github.com/Spirals-Team/docker-machine-driver-g5k
```

Then, either put the driver in a directory filled in your PATH environment variable, or run:

```bash
export PATH=$PATH:$GOPATH/bin
```

Be sure that your SSH public keys file is `$HOME/.ssh/id_rsa.pub`

### Grid5000
You should have an `authorized_keys` file in your SSH directory. Put it on your public directory, on each site you wish to create a machine on.

## How to use

### VPN
You need to be connected to the Grid5000 VPN to create and access your Docker machine node.
Please follow the instructions on the [Grid5000 Wiki](https://www.grid5000.fr/mediawiki/index.php/VPN)

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
