package dev

import (
	"context"
	"fmt"
	"os"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/mage/tools"
	"github.com/aweris/gale/internal/version"
)

type Tools mg.Namespace

// Publish publishes dev version of the tools to the local registry.
func (_ Tools) Publish(ctx context.Context) error {
	registry := os.Getenv("_GALE_DOCKER_REGISTRY")
	if registry == "" {
		return fmt.Errorf("no registry set, please run `mage dev:engine:start` first,than run `eval $(mage dev:engine:env)`")
	}

	v := version.GetVersion()

	return tools.Ghx{}.Publish(ctx, v.GitVersion)
}
