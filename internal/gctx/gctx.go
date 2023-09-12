package gctx

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/data"
)

type Context struct {
	isContainer bool            // isContainer indicates whether the workflow is running in a container.
	debug       bool            // debug indicates whether the workflow is running in debug mode.
	path        string          // path is the data path for the context to be mounted from the host or to be used in the container.
	Context     context.Context // Context is the current context of the workflow.
	Repo        RepoContext     // Repo is the context for the repository.

	// Github Expression Contexts
	Runner  RunnerContext
	Github  GithubContext
	Secrets SecretsContext
	Inputs  InputsContext
	Job     JobContext
	Steps   StepsContext
}

func Load(ctx context.Context, debug bool) (*Context, error) {
	isContainer := os.Getenv(EnvVariableGaleRunner) == "true"

	gctx := &Context{isContainer: isContainer, debug: debug, Context: ctx, path: data.MountPath}

	// load the repository context
	err := gctx.LoadRunnerContext()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadGithubContext()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadSecrets()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadInputs()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadJob()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadSteps()
	if err != nil {
		return nil, err
	}

	// If we can get the token from the environment, we'll use it. Otherwise, we'll try to get it github cli
	if gctx.Github.Token == "" {
		token, err := core.GetToken()
		if err != nil {
			return nil, err
		}

		gctx.SetToken(token)
	} else {
		gctx.Secrets.SetToken(gctx.Github.Token)
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

		// apply sub-contexts
		container = container.With(c.Repo.WithContainerFunc())
		container = container.With(c.Github.WithContainerFunc())
		container = container.With(c.Secrets.WithContainerFunc())
		container = container.With(c.Runner.WithContainerFunc())

		// load repository to container
		container = container.WithMountedDirectory(c.Github.Workspace, c.Repo.Repository.GitRef.Dir)
		container = container.WithWorkdir(c.Github.Workspace)

		return container
	}
}

func (c *Context) SetToken(token string) {
	c.Secrets.SetToken(token)
	c.Github.SetToken(token)
}
