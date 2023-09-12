package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/aweris/gale/internal/cmd"
)

func main() {
	var (
		artifactDir string
		port        string
	)

	pflag.StringVar(&artifactDir, "artifact-dir", "/artifacts", "Directory to store artifacts")
	pflag.StringVar(&port, "port", "8080", "Port to artifact service will listen on")

	cmd.BindEnv(pflag.Lookup("artifact-dir"), "ARTIFACT_DIR")
	cmd.BindEnv(pflag.Lookup("port"), "PORT")

	pflag.Parse()

	if err := Serve(port, NewLocalService(artifactDir)); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}
