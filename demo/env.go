package main

import (
	"fmt"

	"github.com/saschagrunert/demo"
	"github.com/urfave/cli/v2"
)

// env is the environment for the demo to run. It contains helper functions to setup and cleanup the environment and
// build commands to run gale and dagger.
var env = newEnv(BinDir, DaggerVersion, GaleVersion)

// environment contains helper functions to setup and cleanup the environment and build commands to run gale and dagger.
type environment struct {
	dagger dagger // dagger installation information
	gale   gale   // gale installation information
}

// newEnv creates a new environment with the given dagger and gale versions and the directory to install them.
func newEnv(binDir, daggerVersion, galeVersion string) environment {
	return environment{
		dagger: dagger{
			version: daggerVersion,
			dir:     binDir,
		},
		gale: gale{
			version: galeVersion,
			dir:     binDir,
		},
	}
}

// Setup installs dagger and gale
func (e environment) Setup(_ *cli.Context) error {
	fmt.Printf("Installing dagger %s and gale %s...\n", e.dagger.version, e.gale.version)

	return demo.Ensure(e.dagger.Setup(), e.gale.Setup())
}

// RunGaleWithDagger returns a command slice as `dagger run gale run <command>` with correct paths to dagger and gale
func (e environment) RunGaleWithDagger(command string) []string {
	return demo.S(e.dagger.Run(e.gale.Run(command)))
}

// Cleanup removes dagger and gale from the environment
func (e environment) Cleanup(_ *cli.Context) error {
	fmt.Println("Cleaning downloaded binaries...")
	return demo.Ensure(e.dagger.Cleanup(), e.gale.Cleanup())
}

// dagger is a helper struct to configure dagger setup and cleanup commands
type dagger struct {
	version string // version of dagger to install
	dir     string // directory to install dagger
}

// Setup returns the command to install dagger
func (d dagger) Setup() string {
	return "curl -sfLo install_dagger.sh https://releases.dagger.io/dagger/install.sh; DAGGER_VERSION=" + d.version + " BIN_DIR=" + d.dir + " sh ./install_dagger.sh"
}

// Run returns a command as `dagger run <command>` with correct paths to dagger
func (d dagger) Run(command string) string {
	return fmt.Sprintf("%s/dagger run %s", d.dir, command)
}

// Cleanup returns the command to remove dagger
func (d dagger) Cleanup() string {
	return "rm -f ./install_dagger.sh" + d.dir + "/dagger"
}

// gale is a helper struct to configure gale setup and cleanup commands
type gale struct {
	version string // version of gale to install
	dir     string // directory to install gale
}

// Setup returns the command to install gale
func (g gale) Setup() string {
	return "curl -sfLo install_gale.sh https://raw.githubusercontent.com/aweris/gale/main/hack/install.sh; GALE_VERSION=" + g.version + " BIN_DIR=" + g.dir + " sh ./install_gale.sh"
}

// Run returns a command as `gale <command>` with correct paths to gale
func (g gale) Run(command string) string {
	return fmt.Sprintf("%s/gale %s", g.dir, command)
}

// Cleanup returns the command to remove gale
func (g gale) Cleanup() string {
	return "rm -f ./install_gale.sh" + g.dir + "/gale"
}
