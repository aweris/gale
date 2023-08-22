package config

import (
	"path/filepath"

	"dagger.io/dagger"

	"github.com/adrg/xdg"
)

// cfg is the global configuration for ghx. No other package should access it directly.
var cfg = new(config)

func init() {
	cfg.ghxHome = "/home/runner/_temp/ghx"
	// original image is ghcr.io/catthehacker/ubuntu:act-latest. moved to ghcr.io/aweris/gale/runner/ubuntu:22.04
	// to work around issues similar to https://github.com/catthehacker/docker_images/issues/102 and updating the
	// image periodically after testing.
	cfg.runnerImage = "ghcr.io/aweris/gale/runner/ubuntu:22.04"
}

type config struct {
	client      *dagger.Client // client is the dagger client for the config.
	runnerImage string         // runnerImage is the image used for running the actions.
	ghxHome     string         // ghxHome directory where all the data is stored.
}

// SetClient sets the dagger client for the config.
func SetClient(client *dagger.Client) {
	cfg.client = client
}

// Client returns the dagger client for the config.
func Client() *dagger.Client {
	return cfg.client
}

// SetRunnerImage sets the runner image for the config.
func SetRunnerImage(image string) {
	cfg.runnerImage = image
}

// RunnerImage returns the runner image for the config.
func RunnerImage() string {
	return cfg.runnerImage
}

// SetGhxHome sets the ghx home directory for the config.
func SetGhxHome(home string) {
	cfg.ghxHome = home
}

// GhxHome returns the ghx home directory for the config.
func GhxHome() string {
	return cfg.ghxHome
}

// GhxActionsDir returns the directory where the actions are stored.
func GhxActionsDir() string {
	return filepath.Join(GhxHome(), "actions")
}

// GhxRunsDir returns the directory where the runs are stored.
func GhxRunsDir() string {
	return filepath.Join(GhxHome(), "runs")
}

// GhxRunDir returns the directory where the run with the given ID is stored.
func GhxRunDir(runID string) string {
	return filepath.Join(GhxRunsDir(), runID)
}

// GaleDataHome returns the path for local data.
func GaleDataHome() string {
	return filepath.Join(xdg.DataHome, "gale")
}
