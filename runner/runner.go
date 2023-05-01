package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/logger"
)

// Runner represents a GitHub Action runner powered by Dagger.
type Runner struct {
	Client    *dagger.Client
	Container *dagger.Container

	context  *gha.RunContext
	workflow *gha.Workflow
	job      *gha.Job

	ActionsBySource     map[string]*gha.Action
	ActionPathsBySource map[string]string

	log    logger.Logger
	events []Event
}

// NewRunner creates a new Runner.
func NewRunner(ctx context.Context, client *dagger.Client, log logger.Logger, runContext *gha.RunContext, workflow *gha.Workflow, job *gha.Job) (*Runner, error) {
	// check if there is a pre-built runner image
	path, _ := config.SearchDataFile(filepath.Join(config.DefaultRunnerLabel, config.DefaultRunnerImageTar))
	if path != "" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		fmt.Printf("Found pre-built image for %s, importing...\n", config.DefaultRunnerLabel)

		container := client.Container().Import(client.Host().Directory(dir).File(base))

		return &Runner{
			Client:              client,
			Container:           container,
			context:             runContext,
			workflow:            workflow,
			job:                 job,
			ActionsBySource:     make(map[string]*gha.Action),
			ActionPathsBySource: make(map[string]string),
			log:                 log,
		}, nil
	}

	fmt.Printf("No pre-built image found for %s, building a new one...\n", config.DefaultRunnerLabel)

	// Build the runner with the defaults and return it, if there is no pre-built image
	return NewBuilder(client).Build(ctx)
}

// Run runs the job
func (r *Runner) Run(ctx context.Context) {
	r.handle(ctx, SetupJobEvent{})

	for _, step := range r.job.Steps {
		r.ExecStepAction(ctx, "pre", step)
	}

	for _, step := range r.job.Steps {
		r.ExecStepAction(ctx, "main", step)
	}

	for _, step := range r.job.Steps {
		r.ExecStepAction(ctx, "post", step)
	}
}
