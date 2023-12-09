package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v9"
)

// ServiceConfig is the configuration for the artifactcache service.
type ServiceConfig struct {
	CacheDir string `env:"CACHE_DIR" envDefault:"/cache"`
	Port     string `env:"PORT" envDefault:"8080"`
}

func main() {
	var config ServiceConfig

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Error parsing environment variables: %s\n", err.Error())
		os.Exit(1)
	}

	srv, err := NewLocalService(config.CacheDir)
	if err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}

	if err := Serve(config.Port, srv); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}
