package config

import "dagger.io/dagger"

// cfg is the global configuration for ghx. No other package should access it directly.
var cfg = new(config)

type config struct {
	client *dagger.Client
}

// SetClient sets the dagger client for the config.
func SetClient(client *dagger.Client) {
	cfg.client = client
}

// Client returns the dagger client for the config.
func Client() *dagger.Client {
	return cfg.client
}
