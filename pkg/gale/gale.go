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
		container = container.With(core.NewGithubRepositoryContext(g.repo).WithContainerFunc())
		container = container.With(core.NewGithubURLContext().WithContainerFunc())
		container = container.With(core.NewGithubRefContext(g.repo.GitRef).WithContainerFunc())

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

		/*
			// FIXME: move this to ghx

			//TODO: not sure if this is the best way to get the sha, probably we're missing some scenarios. However, it works for now.

			sha, err := config.Client().Container().
				From("alpine/git").
				WithMountedDirectory("/workdir", repo.GitRef.Dir).
				WithWorkdir("/workdir").
				WithExec([]string{"log", "-1", "--follow", "--format=%H", "--", wf.Path}).
				Stdout(ctx)
			if err != nil {
				return helpers.FailPipeline(container, err)
			}

		*/

		// context configuration
		container = container.With(core.NewGithubSecretsContext(token).WithContainerFunc())
		// FIXME: move this to ghx
		// container = container.With(core.NewGithubWorkflowContext(repo, wf, workflowRunID, strings.TrimSpace(sha)).WithContainerFunc())
		// container = container.With(core.NewGithubJobInfoContext(job).WithContainerFunc())
		container = container.With(core.NewGithubEventContext().WithContainerFunc())
		container = container.With(core.NewSecretsContext(token, opt.Secrets).WithContainerFunc())

		container = container.WithEnvVariable("GALE_WORKFLOWS_DIR", opt.WorkflowsDir)
		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", workflow, job})

		return container
	}
}
