package images

import (
	"runtime"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
)

// RunnerBase returns a container with the base image for the runner.
func RunnerBase() *dagger.Container {
	return config.Client().Container().
		Pipeline("Runner Base Image").
		From("ghcr.io/catthehacker/ubuntu:act-22.04")
}

// GoBase returns a container with the base image for the go
func GoBase() *dagger.Container {
	return config.Client().Container().
		Pipeline("Go Base Image").
		From("golang:" + strings.TrimPrefix(runtime.Version(), "go"))
}
