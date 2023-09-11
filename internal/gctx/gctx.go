package gctx

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/data"
)

type Context struct {
	isContainer bool            // isContainer indicates whether the workflow is running in a container.
	path        string          // path is the data path for the context to be mounted from the host or to be used in the container.
	Context     context.Context // Context is the current context of the workflow.
	Repo        RepoContext     // Repo is the context for the repository.

	// Github Contexts
	Secret SecretsContext
}

func Load(ctx context.Context) (*Context, error) {
	isContainer := os.Getenv(EnvVariableGaleRunner) == "true"

	gctx := &Context{isContainer: isContainer, Context: ctx, path: data.MountPath}

	if isContainer {
		gctx.LoadSecrets()
	}

	return gctx, nil
}

var _ helpers.WithContainerFuncHook = new(Context)

func (c *Context) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// set the environment variable that indicates that the workflow is running in a container.
		// using this variable, we can distinguish between the container and the host process and configure the
		// context accordingly.
		container = container.WithEnvVariable(EnvVariableGaleRunner, "true")

		// apply sub contexts to the container
		container = c.Repo.WithContainerFunc()(container)
		container = c.Secret.WithContainerFunc()(container)

		return container
	}
}
