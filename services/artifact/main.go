package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/pflag"
)

func main() {
	var (
		artifactDir string
		port        string
	)

	pflag.StringVar(&artifactDir, "artifact-dir", "/artifacts", "Directory to store artifacts")
	pflag.StringVar(&port, "port", "8080", "Port to artifact service will listen on")

	bindEnv(pflag.Lookup("artifact-dir"), "ARTIFACT_DIR")
	bindEnv(pflag.Lookup("port"), "PORT")

	pflag.Parse()

	if err := Serve(port, NewLocalService(artifactDir)); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
