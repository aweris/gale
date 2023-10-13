package config

import (
	"os"
	"path/filepath"
)

// FIXME: global config was a mistake. It's hard to maintain and track where it's used. We should refactor this to
//  something more maintainable.

// cfg is the global configuration for ghx. No other package should access it directly.
var cfg = new(config)

func init() {
	cfg.ghxHome = "/home/runner/_temp/ghx"
}

// TODO: need some change here to make it more maintainable. This became a hacky mess and it's hard to maintain or
//  track where it's used.

type config struct {
	ghxHome string // ghxHome directory where all the data is stored.
	debug   bool   // debug is the flag to enable debug mode.
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

// Debug returns the debug flag for the config. It also checks the environment variable RUNNER_DEBUG to enable debug
// mode. If RUNNER_DEBUG=1 is set, it will enable debug mode.
func Debug() bool {
	return cfg.debug || os.Getenv("RUNNER_DEBUG") == "1"
}
