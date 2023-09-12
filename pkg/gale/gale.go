package gale

import (
	"context"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/data"
)

// RunOpts are the options for the Run function.
type RunOpts struct {
	WorkflowsDir string
	Secrets      map[string]string
}

// Gale is the main entrypoint for the gale library.
type Gale struct {
	repo             *core.Repository
	ghx              *Ghx
	artifactSVC      *ArtifactService
	artifactCacheSVC *ArtifactCacheService
}

// New creates a new gale instance.
func New(repo *core.Repository) *Gale {
	cache := data.NewCacheVolume(repo)

	return &Gale{
		repo:             repo,
		ghx:              NewGhxBinary(cache),
		artifactSVC:      NewArtifactService(cache),
		artifactCacheSVC: NewArtifactCacheService(cache),
	}
}

// ExecutionEnv returns a dagger function that sets the execution environment of the gale to the given container.
func (g *Gale) ExecutionEnv(_ context.Context) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// pass dagger context to the container
		container = container.With(core.NewDaggerContextFromEnv().WithContainerFunc())

		// tools
		container = container.With(g.ghx.WithContainerFunc())

		// services
		container = container.With(g.artifactSVC.WithContainerFunc())
		container = container.With(g.artifactCacheSVC.WithContainerFunc())

		// context configuration -- these are the contexts that not change during the execution
		container = container.With(core.NewRunnerContext(config.Debug()).WithContainerFunc())

		return container
	}
}

// Run runs a job from a workflow.
func (g *Gale) Run(_ context.Context, workflow, job string, opts ...RunOpts) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		opt := RunOpts{}
		if len(opts) > 0 {
			opt = opts[0]
		}

		token, err := core.GetToken()
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		// context configuration
		gc := core.NewGithubContext(g.repo, token)

		container = container.With(gc.WithContainerFunc())
		container = container.With(core.NewSecretsContext(token, opt.Secrets).WithContainerFunc())

		// load repository to container
		container = container.WithMountedDirectory(gc.Workspace, g.repo.GitRef.Dir)
		container = container.WithWorkdir(gc.Workspace)

		container = container.WithEnvVariable("GALE_WORKFLOWS_DIR", opt.WorkflowsDir)
		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", workflow, job})

		return container
	}
}
