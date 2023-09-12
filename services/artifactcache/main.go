package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/aweris/gale/internal/cmd"
)

func main() {
	var (
		cacheDir         string
		port             string
		externalHostname string
	)

	pflag.StringVar(&cacheDir, "cache-dir", "/cache", "Directory to store cache")
	pflag.StringVar(&port, "port", "8080", "Port to artifact service will listen on")
	pflag.StringVar(&externalHostname, "external-hostname", "", "External hostname to use for download URLs")

	cmd.BindEnv(pflag.Lookup("cache-dir"), "CACHE_DIR")
	cmd.BindEnv(pflag.Lookup("port"), "PORT")
	cmd.BindEnv(pflag.Lookup("external-hostname"), "EXTERNAL_HOSTNAME")

	pflag.Parse()

	srv, err := NewLocalService(cacheDir, externalHostname, port)
	if err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}

	if err := Serve(port, srv); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}
