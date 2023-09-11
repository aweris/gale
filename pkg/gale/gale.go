package gale

import (
	"context"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/gctx"
)

// Gale is the main entrypoint for the gale library.
type Gale struct {
	rc               *gctx.Context
	ghx              *Ghx
	artifactSVC      *ArtifactService
	artifactCacheSVC *ArtifactCacheService
}

// New creates a new gale instance.
func New(rc *gctx.Context) *Gale {
	return &Gale{
		rc:               rc,
		ghx:              NewGhxBinary(rc.Repo.CacheVol),
		artifactSVC:      NewArtifactService(rc.Repo.CacheVol),
		artifactCacheSVC: NewArtifactCacheService(rc.Repo.CacheVol),
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
func (g *Gale) Run(_ context.Context, workflow, job string) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		token, err := core.GetToken()
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		//TODO: refactor later
		g.rc.Secret.SetToken(token)

		// context configuration
		container = container.With(g.rc.WithContainerFunc())

		gc := core.NewGithubContext(g.rc.Repo.Repository, token)

		container = container.With(gc.WithContainerFunc())

		// load repository to container
		container = container.WithMountedDirectory(gc.Workspace, g.rc.Repo.Repository.GitRef.Dir)
		container = container.WithWorkdir(gc.Workspace)

		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", workflow, job})

		return container
	}
}
