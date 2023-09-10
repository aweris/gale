package config

import (
	"os"
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

// TODO: need some change here to make it more maintainable. This became a hacky mess and it's hard to maintain or
//  track where it's used.

type config struct {
	client      *dagger.Client // client is the dagger client for the config.
	clientNoLog *dagger.Client // clientNoLog is the dagger client without logging.
	runnerImage string         // runnerImage is the image used for running the actions.
	ghxHome     string         // ghxHome directory where all the data is stored.
	debug       bool           // debug is the flag to enable debug mode.
}

// SetClient sets the dagger client for the config.
func SetClient(client *dagger.Client) {
	cfg.client = client
}

// Client returns the dagger client for the config.
func Client() *dagger.Client {
	return cfg.client
}

// SetClientNoLog sets the dagger client without logging for the config.
func SetClientNoLog(client *dagger.Client) {
	cfg.clientNoLog = client
}

// ClientNoLog returns the dagger client without logging for the config.
func ClientNoLog() *dagger.Client {
	return cfg.clientNoLog
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

// SetDebug sets the debug flag for the config.
func SetDebug(debug bool) {
	cfg.debug = debug
}

// Debug returns the debug flag for the config. It also checks the environment variable RUNNER_DEBUG to enable debug
// mode. If RUNNER_DEBUG=1 is set, it will enable debug mode.
func Debug() bool {
	return cfg.debug || os.Getenv("RUNNER_DEBUG") == "1"
}
