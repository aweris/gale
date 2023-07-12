//go:build mage

package main

import (
	//mage:import services
	_ "github.com/aweris/gale/internal/mage/services"

	//mage:import tools
	_ "github.com/aweris/gale/internal/mage/tools"
)
