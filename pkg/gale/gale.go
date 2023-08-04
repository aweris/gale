package gale

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/dagger/services"
	"github.com/aweris/gale/internal/dagger/tools"
	"github.com/aweris/gale/internal/idgen"
)

// RunOpts are the options for the Run function.
type RunOpts struct {
	Repo         string
	Branch       string
	Tag          string
	Commit       string
	WorkflowsDir string
}

// Run runs a job from a workflow.
func Run(ctx context.Context, workflow, job string, opts ...RunOpts) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		opt := RunOpts{}
		if len(opts) > 0 {
			opt = opts[0]
		}

		repo, err := core.GetRepository(opt.Repo, core.GetRepositoryOpts{Branch: opt.Branch, Tag: opt.Tag, Commit: opt.Commit})
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		workflows, err := repo.LoadWorkflows(ctx, core.RepositoryLoadWorkflowOpts{WorkflowsDir: opt.WorkflowsDir})
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		wf, ok := workflows[workflow]
		if !ok {
			return helpers.FailPipeline(container, ErrWorkflowNotFound)
		}

		jm, ok := wf.Jobs[job]
		if !ok {
			return helpers.FailPipeline(container, ErrJobNotFound)
		}

		// ensure job name is set
		if jm.Name == "" {
			jm.Name = job
		}

		workflowRunID, err := idgen.GenerateWorkflowRunID(repo)
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		jobRunID, err := idgen.GenerateJobRunID(repo)
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		jr := &core.JobRun{
			RunID: jobRunID,
			Job:   jm,
		}

		dir, err := core.MarshalJobRunToDir(ctx, jr)
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		token, err := core.GetToken()
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		// context configuration
		container = container.With(core.NewRunnerContext().Apply)
		container = container.With(core.NewGithubRepositoryContext(repo).Apply)
		container = container.With(core.NewGithubSecretsContext(token).Apply)
		container = container.With(core.NewGithubURLContext().Apply)
		container = container.With(core.NewGithubWorkflowContext(repo, wf, workflowRunID).Apply)
		container = container.With(core.NewGithubJobInfoContext(job).Apply)

		// job run configuration
		container = container.WithDirectory(config.GhxRunDir(jobRunID), dir)
		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", jobRunID})

		return container
	}
}

// ExecutionEnv returns a dagger function that sets the execution environment of the gale to the given container.
func ExecutionEnv(_ context.Context) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// pass dagger context to the container
		container = container.With(core.NewDaggerContextFromEnv().Apply)

		// tools

		ghx, err := tools.Ghx(context.Background())
		if err != nil {
			helpers.FailPipeline(container, fmt.Errorf("error getting ghx: %w", err))
		}

		container = container.WithFile("/usr/local/bin/ghx", ghx)
		container = container.WithEnvVariable("GHX_HOME", config.GhxHome())
		container = container.WithMountedCache(config.GhxActionsDir(), config.Client().CacheVolume("actions"))

		// services

		container = container.With(services.NewArtifactService().ServiceBinding)

		return container
	}
}
