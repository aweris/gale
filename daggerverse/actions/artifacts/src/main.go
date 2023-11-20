package main

import (
	"fmt"
	"os"

	"github.com/caarlos0/env/v9"
)

type ServiceConfig struct {
	ArtifactDir string `env:"ARTIFACT_DIR" envDefault:"/artifacts"`
	Port        string `env:"PORT" envDefault:"8080"`
}

func main() {
	var config ServiceConfig

	if err := env.Parse(&config); err != nil {
		fmt.Printf("Error parsing environment variables: %s\n", err.Error())
		os.Exit(1)
	}

	if err := Serve(config.Port, NewLocalService(config.ArtifactDir)); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}
