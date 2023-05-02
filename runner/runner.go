package runner

import (
	"context"
	"dagger.io/dagger"
	"github.com/aweris/gale/config"
	"path/filepath"
	"sync/atomic"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/logger"
)

var _ Runner = new(runner)

type Runner interface {
	Run(ctx context.Context)
}

// runner represents a GitHub Action runner powered by Dagger.
type runner struct {
	client    *dagger.Client
	container *dagger.Container

	context  *gha.RunContext
	workflow *gha.Workflow
	job      *gha.Job

	actionsBySource     map[string]*gha.Action
	actionPathsBySource map[string]string

	log     logger.Logger
	events  []*EventRecord
	counter *atomic.Uint64
}

// NewRunner creates a new Runner.
func NewRunner(client *dagger.Client, log logger.Logger, runContext *gha.RunContext, workflow *gha.Workflow, job *gha.Job) Runner {
	return &runner{
		client:              client,
		context:             runContext,
		workflow:            workflow,
		job:                 job,
		actionsBySource:     make(map[string]*gha.Action),
		actionPathsBySource: make(map[string]string),
		log:                 log,
		counter:             &atomic.Uint64{},
	}
}

// Run runs the job
func (r *runner) Run(ctx context.Context) {
	path, _ := config.SearchDataFile(filepath.Join(config.DefaultRunnerLabel, config.DefaultRunnerImageTar))

	// Load or build container
	if path != "" {
		r.handle(ctx, LoadContainerEvent{Path: path})
	} else {
		r.handle(ctx, BuildContainerEvent{})
	}

	// Setup Job
	r.handle(ctx, SetupJobEvent{})

	// Run stages
	for _, step := range r.job.Steps {
		r.handle(ctx, ExecStepActionEvent{Stage: "pre", Step: step})
	}

	for _, step := range r.job.Steps {
		r.handle(ctx, ExecStepActionEvent{Stage: "main", Step: step})
	}

	for _, step := range r.job.Steps {
		r.handle(ctx, ExecStepActionEvent{Stage: "post", Step: step})
	}
}
