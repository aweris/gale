//go:build mage

package main

import (
	//mage:import docker
	_ "github.com/aweris/gale/common/mage/docker"

	//mage:import dev
	_ "github.com/aweris/gale/common/mage/dev"
)
