package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/mage/docker"
	"github.com/aweris/gale/internal/version"
)

type Docker mg.Namespace

// Publish publishes dev version of the docker image to the local registry.
func (_ Docker) Publish(ctx context.Context) error {
	registry := os.Getenv("_GALE_DOCKER_REGISTRY")
	if registry == "" {
		return fmt.Errorf("no registry set, please run `mage dev:engine:start` first,than run `eval $(mage dev:engine:env)`")
	}

	v := version.GetVersion()

	return docker.Publish(ctx, v.GitVersion)
}
