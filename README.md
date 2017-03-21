[![Build Status](https://travis-ci.org/Spirals-Team/docker-machine-driver-g5k.svg)](https://travis-ci.org/Spirals-Team/docker-machine-driver-g5k)

# docker-machine-driver-g5k
A Docker Machine driver for the Grid5000 testbed infrastructure. It can be used to provision a Docker machine on a node of the Grid5000 infrastructure.

## Requirements
* [Docker](https://www.docker.com/products/overview#/install_the_platform)
* [Docker Machine](https://docs.docker.com/machine/install-machine)
* [Go tools (Only for installation from sources)](https://golang.org/doc/install)

You need a Grid5000 account to use this driver. See [this page](https://www.grid5000.fr/mediawiki/index.php/Grid5000:Get_an_account) to create an account.

## Installation from GitHub releases
Binary releases for Linux, MacOS and Windows using x86/x86_64 CPU architectures are available in the [releases page](https://github.com/Spirals-Team/docker-machine-driver-g5k/releases).
You can use the following commands to install or upgrade the driver:
```bash
# download the binary for your OS and CPU architecture :
sudo curl -L -o /usr/local/bin/docker-machine-driver-g5k "<link to release>"

# grant execution rigths to the driver for everyone :
sudo chmod +x /usr/local/bin/docker-machine-driver-g5k
```

## Installation from sources
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
Please follow the instructions from the [Grid5000 Wiki](https://www.grid5000.fr/mediawiki/index.php/VPN).

### Driver-specific options
The driver needs a few options to create a machine. Here is a list of the supported options:

|            Option            |                       Description                       |     Default value     |  Required  |
|------------------------------|---------------------------------------------------------|-----------------------|------------|
| `--g5k-username`             | Your Grid5000 account username                          |                       | Yes        |
| `--g5k-password`             | Your Grid5000 account password                          |                       | Yes        |
| `--g5k-site`                 | Site to reserve the resources on                        |                       | Yes        |
| `--g5k-walltime`             | Timelife of the machine                                 | "1:00:00"             | No         |
| `--g5k-image`                | Name of the image to deploy                             | "jessie-x64-min"      | No         |
| `--g5k-resource-properties`  | Resource selection with OAR properties (SQL format)     |                       | No         |
| `--g5k-use-job-reservation`  | Job ID to use (need to be an already existing job ID)   |                       | No         |
| `--g5k-host-to-provision`    | Host to provision (host need to be already deployed)    |                       | No         |

#### Resource properties
You can use [OAR properties](http://oar.imag.fr/docs/2.5/user/usecases.html#using-properties) to only select a node that matches your hardware requirements.  
If you give incorrect properties or no resource matches your request, you will get this error:
```bash
Error with pre-create check: "[G5K_api] request failed: 400 Bad Request."
```

More informations about usage of OAR properties are available on the [Grid5000 Wiki](https://www.grid5000.fr/mediawiki/index.php/Advanced_OAR#Other_examples_using_properties).

### Provisioning examples
An example of node provisioning:
```bash
docker-machine create -d g5k \
--g5k-username "user" \
--g5k-password "********" \
--g5k-site "lille" \
test-node
```

An example of node provisioning using environment variables:
```bash
export G5K_USERNAME="user"
export G5K_PASSWORD="********"
export G5K_SITE="lille"
docker-machine create -d g5k test-node
```

An example with resource properties (node in cluster `chimint` with more thant 8GB of RAM and at least 4 CPU cores):
```bash
docker-machine create -d g5k \
--g5k-username "user" \
--g5k-password "********" \
--g5k-site "lille" \
--g5k-resource-properties "cluster = 'chimint' and memnode > 8192 and cpucore >= 4" \
test-node
```

An example using an existing oarsub job ID and a host already deployed with kadeploy3:
```bash
docker-machine create -d g5k \
--g5k-username "user" \
--g5k-password "********" \
--g5k-site "lille" \
--g5k-use-job-reservation 1234567 \
--g5k-host-to-provision "chinqchint-xx.lille.grid5000.fr" \
test-node
``` 
