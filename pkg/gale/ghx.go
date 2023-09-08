package gale

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
	"github.com/aweris/gale/pkg/data"
)

var _ helpers.WithContainerFuncHook = new(Ghx)

type Ghx struct {
	tag  string
	cli  *dagger.File
	data *data.CacheVolume
}

// NewGhxBinary returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func NewGhxBinary(data *data.CacheVolume) *Ghx {
	v := version.GetVersion()

	tag := v.GitVersion

	file := config.Client().Container().From("ghcr.io/aweris/gale/tools/ghx:" + tag).File("/ghx")

	return &Ghx{tag: tag, cli: file, data: data}
}

func (g *Ghx) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// check, if the file doesn't exist or is empty
		if size, err := g.cli.Size(context.Background()); size == 0 || err != nil {
			return helpers.FailPipeline(container, fmt.Errorf("ghx@%s binary not available", g.tag))
		}

		// add the ghx binary to the container
		container = container.WithFile("/usr/local/bin/ghx", g.cli)

		// add the gale data directory to the container
		container = container.With(g.data.WithContainerFunc()).WithEnvVariable("GHX_HOME", data.MountPath)

		return container
	}
}
