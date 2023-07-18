package config

import (
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/tools/ghx/internal/log"
)

// cfg is the global configuration for ghx. No other package should access it directly.
var cfg = config{
	home:   "/home/runner/_temp/ghx",
	logger: log.NewLogger(),
}

// config represents the configuration for ghx.
type config struct {
	home   string
	client *dagger.Client
	logger *log.Logger
}

// SetConfigHome sets the home directory for the ghx configuration.
func SetConfigHome(home string) {
	cfg.home = home
}

// SetClient sets the dagger client for the ghx configuration.
func SetClient(client *dagger.Client) {
	cfg.client = client
}

// Home returns the home directory for the ghx configuration.
func Home() string {
	return cfg.home
}

// ActionsDir returns the directory where the actions are stored.
func ActionsDir() string {
	return filepath.Join(cfg.home, "actions")
}

// RunsDir returns the directory where the runs are stored.
func RunsDir() string {
	return filepath.Join(cfg.home, "runs")
}

// RunDir returns the directory where the run with the given ID is stored.
func RunDir(runID string) string {
	return filepath.Join(RunsDir(), runID)
}

func Client() *dagger.Client {
	return cfg.client
}

func Logger() *log.Logger {
	return cfg.logger
}
