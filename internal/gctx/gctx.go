package gctx

import (
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/dagger/helpers"
)

type Context struct {
	isContainer bool // isContainer indicates whether the workflow is running in a container.
}

func Load() (*Context, error) {
	isContainer := os.Getenv(EnvVariableGaleRunner) == "true"

	gctx := &Context{
		isContainer: isContainer,
	}

	return gctx, nil
}

var _ helpers.WithContainerFuncHook = new(Context)

func (c Context) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// set the environment variable that indicates that the workflow is running in a container.
		// using this variable, we can distinguish between the container and the host process and configure the
		// context accordingly.
		container = container.WithEnvVariable(EnvVariableGaleRunner, "true")

		return container
	}
}
