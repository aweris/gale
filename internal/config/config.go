package config

import "dagger.io/dagger"

// cfg is the global configuration for ghx. No other package should access it directly.
var cfg = new(config)

func init() {
	cfg.ghxHome = "/home/runner/_temp/ghx"
}

type config struct {
	client  *dagger.Client // dagger client for the config.
	ghxHome string         // ghx home directory where all the data is stored.
}

// SetClient sets the dagger client for the config.
func SetClient(client *dagger.Client) {
	cfg.client = client
}

// Client returns the dagger client for the config.
func Client() *dagger.Client {
	return cfg.client
}

// SetGhxHome sets the ghx home directory for the config.
func SetGhxHome(home string) {
	cfg.ghxHome = home
}

// GhxHome returns the ghx home directory for the config.
func GhxHome() string {
	return cfg.ghxHome
}
