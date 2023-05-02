package runner

import (
	"context"
	"dagger.io/dagger"
	"github.com/aweris/gale/config"
	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
	"github.com/aweris/gale/logger"
	"path/filepath"
)

var _ Runner = new(runner)

type Runner interface {
	Run(ctx context.Context)
}

// runner represents a GitHub Action runner powered by Dagger.
type runner struct {
	context   *Context
	publisher event.Publisher[Context]
}

var _ event.Context = new(Context)

type Context struct {
	client    *dagger.Client
	container *dagger.Container

	context  *gha.RunContext
	workflow *gha.Workflow
	job      *gha.Job

	actionsBySource     map[string]*gha.Action
	actionPathsBySource map[string]string

	log logger.Logger
}

// NewRunner creates a new Runner.
func NewRunner(client *dagger.Client, log logger.Logger, runContext *gha.RunContext, workflow *gha.Workflow, job *gha.Job) Runner {
	rc := &Context{
		client:              client,
		context:             runContext,
		workflow:            workflow,
		job:                 job,
		actionsBySource:     make(map[string]*gha.Action),
		actionPathsBySource: make(map[string]string),
		log:                 log,
	}
	return &runner{
		context:   rc,
		publisher: event.NewStdPublisher(rc),
	}
}

// Run runs the job
func (r *runner) Run(ctx context.Context) {
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
		r.publisher.Publish(ctx, ExecStepActionEvent{Stage: "pre", Step: step})
	}

	for _, step := range r.context.job.Steps {
		r.publisher.Publish(ctx, ExecStepActionEvent{Stage: "main", Step: step})
	}

	for _, step := range r.context.job.Steps {
		r.publisher.Publish(ctx, ExecStepActionEvent{Stage: "post", Step: step})
	}
}
