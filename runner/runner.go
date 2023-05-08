package runner

import (
	"context"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
	"github.com/aweris/gale/github/actions"
	"github.com/aweris/gale/internal/event"
	"github.com/aweris/gale/logger"
)

var _ Runner = new(runner)

type Runner interface {
	Run(ctx context.Context, rc *actions.RunContext, workflow *actions.Workflow, job *actions.Job)
}

// runner represents a GitHub Action runner powered by Dagger.
type runner struct {
	context   *Context
	publisher event.Publisher[Context]
}

// NewRunner creates a new Runner.
func NewRunner(client *dagger.Client, log logger.Logger) Runner {
	rc := NewContext(client, log)

	return &runner{context: rc, publisher: event.NewStdPublisher(rc)}
}

// Run runs the job
func (r *runner) Run(ctx context.Context, runContext *actions.RunContext, workflow *actions.Workflow, job *actions.Job) {
	// update context with new run
	r.context.context = runContext
	r.context.workflow = workflow
	r.context.job = job

	path, _ := config.SearchDataFile(filepath.Join(config.DefaultRunnerLabel, config.DefaultRunnerImageTar))

	// Load or build container
	if path != "" {
		r.publisher.Publish(ctx, LoadContainerEvent{Path: path})
	} else {
		r.publisher.Publish(ctx, BuildContainerEvent{})
	}

	// Setup Job
	r.publisher.Publish(ctx, SetupJobEvent{})

	// Run stages
	for _, step := range r.context.job.Steps {
		r.publisher.Publish(ctx, ExecStepEvent{Stage: actions.ActionStagePre, Step: step})
	}

	for _, step := range r.context.job.Steps {
		r.publisher.Publish(ctx, ExecStepEvent{Stage: actions.ActionStageMain, Step: step})
	}

	for _, step := range r.context.job.Steps {
		r.publisher.Publish(ctx, ExecStepEvent{Stage: actions.ActionStagePost, Step: step})
	}
}
