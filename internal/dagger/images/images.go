package images

import (
	"runtime"
	"strings"

	"dagger.io/dagger"
)

// RunnerBase returns a container with the base image for the runner.
func RunnerBase(client *dagger.Client) *dagger.Container {
	return client.Container().
		Pipeline("Runner Base Image").
		From("ghcr.io/catthehacker/ubuntu:act-22.04")
}

// GoBase returns a container with the base image for the go
func GoBase(client *dagger.Client) *dagger.Container {
	return client.Container().
		Pipeline("Go Base Image").
		From("golang:" + strings.TrimPrefix(runtime.Version(), "go"))
}
