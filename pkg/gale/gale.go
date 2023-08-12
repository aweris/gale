package gale

import (
	"context"
	"strings"

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
	WorkflowsDir string
	Secrets      map[string]string
}

// Run runs a job from a workflow.
func Run(ctx context.Context, workflow, job string, opts ...RunOpts) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		opt := RunOpts{}
		if len(opts) > 0 {
			opt = opts[0]
		}

		repo, err := core.GetRepository(opt.Repo, core.GetRepositoryOpts{Branch: opt.Branch, Tag: opt.Tag})
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

		token, err := core.GetToken()
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

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

		// context configuration
		container = container.With(core.NewRunnerContext().WithContainerFunc())
		container = container.With(core.NewGithubRepositoryContext(repo).WithContainerFunc())
		container = container.With(core.NewGithubSecretsContext(token).WithContainerFunc())
		container = container.With(core.NewGithubURLContext().WithContainerFunc())
		container = container.With(core.NewGithubWorkflowContext(repo, wf, workflowRunID, strings.TrimSpace(sha)).WithContainerFunc())
		container = container.With(core.NewGithubJobInfoContext(job).WithContainerFunc())
		container = container.With(core.NewGithubRefContext(repo.GitRef).WithContainerFunc())
		container = container.With(core.NewSecretsContext(token, opt.Secrets).WithContainerFunc())

		// job run configuration
		container = container.With(core.NewJobRun(jobRunID, jm).WithContainerFunc())

		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", jobRunID})

		return container
	}
}

// ExecutionEnv returns a dagger function that sets the execution environment of the gale to the given container.
func ExecutionEnv(_ context.Context) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// pass dagger context to the container
		container = container.With(core.NewDaggerContextFromEnv().WithContainerFunc())

		// tools
		container = container.With(tools.NewGhxBinary().WithContainerFunc())

		// services
		container = container.With(services.NewArtifactService().WithContainerFunc()) // TODO: move service to context or outside to be able to use it later to get artifacts

		return container
	}
}
