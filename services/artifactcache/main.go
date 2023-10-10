package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/aweris/gale/internal/cmd"
)

func main() {
	var (
		cacheDir string
		port     string
	)

	pflag.StringVar(&cacheDir, "cache-dir", "/cache", "Directory to store cache")
	pflag.StringVar(&port, "port", "8080", "Port to artifact service will listen on")

	cmd.BindEnv(pflag.Lookup("cache-dir"), "CACHE_DIR")
	cmd.BindEnv(pflag.Lookup("port"), "PORT")

	pflag.Parse()

	srv, err := NewLocalService(cacheDir)
	if err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}

	if err := Serve(port, srv); err != nil {
		fmt.Printf("Error starting artifact service: %s\n", err.Error())
		os.Exit(1)
	}
}
