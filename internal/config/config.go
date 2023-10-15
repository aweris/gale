package config

import (
	"os"
)

// Debug returns the debug flag for the config. It also checks the environment variable RUNNER_DEBUG to enable debug
// mode. If RUNNER_DEBUG=1 is set, it will enable debug mode.
func Debug() bool {
	return os.Getenv("RUNNER_DEBUG") == "1"
}
