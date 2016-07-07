package main

import (
    "docker-machine-driver-g5k/driver"
    "github.com/docker/machine/libmachine/drivers/plugin"
)

func main() {
    plugin.RegisterDriver(driver.NewDriver())
}
