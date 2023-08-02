package gale

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/google/uuid"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/tools"
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
			return fail(container, err)
		}

		workflows, err := repo.LoadWorkflows(ctx, core.RepositoryLoadWorkflowOpts{WorkflowsDir: opt.WorkflowsDir})
		if err != nil {
			return fail(container, err)
		}

		wf, ok := workflows[workflow]
		if !ok {
			return fail(container, ErrWorkflowNotFound)
		}

		jm, ok := wf.Jobs[job]
		if !ok {
			return fail(container, ErrJobNotFound)
		}

		// ensure job name is set
		if jm.Name == "" {
			jm.Name = job
		}

		runID := uuid.New().String()

		jr := &core.JobRun{
			RunID: runID,
			Job:   jm,
		}

		dir, err := core.MarshalJobRunToDir(ctx, jr)
		if err != nil {
			return fail(container, err)
		}

		token, err := core.GetToken()
		if err != nil {
			return fail(container, err)
		}

		// context configuration
		container = container.With(core.NewRunnerContext().Apply)
		container = container.With(core.NewGithubRepositoryContext(repo).Apply)
		container = container.With(core.NewGithubSecretsContext(token).Apply)
		container = container.With(core.NewGithubURLContext().Apply)
		container = container.With(core.NewGithubWorkflowContext(repo, wf).Apply)
		container = container.With(core.NewGithubJobInfoContext(job).Apply)

		// job run configuration
		container = container.WithDirectory(config.GhxRunDir(runID), dir)
		container = container.WithExec([]string{"/usr/local/bin/ghx", "run", runID})

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
			fail(container, fmt.Errorf("error getting ghx: %w", err))
		}

		container = container.WithFile("/usr/local/bin/ghx", ghx)
		container = container.WithEnvVariable("GHX_HOME", config.GhxHome())
		container = container.WithMountedCache(config.GhxActionsDir(), config.Client().CacheVolume("actions"))

		return container
	}
}

// fail returns a container that immediately fails with the given error. This useful for forcing a pipeline to fail
// inside chaining operations.
func fail(container *dagger.Container, err error) *dagger.Container {
	// fail the container with the given error
	container = container.WithExec([]string{"sh", "-c", "echo " + err.Error() + " && exit 1"})

	// forced evaluation of the pipeline to immediately fail
	container, _ = container.Sync(context.Background())

	return container
}
