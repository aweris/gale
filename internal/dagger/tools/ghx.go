package tools

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
)

var _ helpers.WithContainerFuncHook = new(Ghx)

type Ghx struct {
	tag  string
	file *dagger.File
}

// NewGhxBinary returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func NewGhxBinary() *Ghx {
	v := version.GetVersion()

	tag := v.GitVersion

	file := config.Client().Container().From("ghcr.io/aweris/gale/tools/ghx:" + tag).File("/ghx")

	return &Ghx{tag: tag, file: file}
}

func (g *Ghx) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// check, if the file doesn't exist or is empty
		if size, err := g.file.Size(context.Background()); size == 0 || err != nil {
			return helpers.FailPipeline(container, fmt.Errorf("ghx@%s binary not available", g.tag))
		}

		return container.WithFile("/usr/local/bin/ghx", g.file).
			WithEnvVariable("GHX_HOME", config.GhxHome()).
			WithMountedCache(config.GhxActionsDir(), config.Client().CacheVolume("actions"))
	}
}
