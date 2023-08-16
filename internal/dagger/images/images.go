package images

import (
	"runtime"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
)

// RunnerBase returns a container with the base image for the runner.
func RunnerBase() *dagger.Container {
	// original image is ghcr.io/catthehacker/ubuntu:act-latest. moved to ghcr.io/aweris/gale/runner/ubuntu:22.04
	// to work around issues similar to https://github.com/catthehacker/docker_images/issues/102 and updating the
	// image periodically after testing.
	return config.Client().Container().From("ghcr.io/aweris/gale/runner/ubuntu:22.04")
}

// GoBase returns a container with the base image for the go
func GoBase() *dagger.Container {
	return config.Client().Container().From("golang:" + strings.TrimPrefix(runtime.Version(), "go"))
}
