package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/mage/services"
	"github.com/aweris/gale/internal/version"
)

type Services mg.Namespace

// Publish publishes dev version of the services to the local registry.
func (_ Services) Publish(ctx context.Context) error {
	registry := os.Getenv("_GALE_DOCKER_REGISTRY")
	if registry == "" {
		return fmt.Errorf("no registry set, please run `mage dev:engine:start` first,than run `eval $(mage dev:engine:env)`")
	}

	v := version.GetVersion()

	err := services.Artifact{}.Publish(ctx, v.GitVersion)
	if err != nil {
		return err
	}

	err = services.ArtifactCache{}.Publish(ctx, v.GitVersion)
	if err != nil {
		return err
	}

	return nil
}
