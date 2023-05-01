package executor

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/logger"
	runnerpkg "github.com/aweris/gale/runner"
)

type JobExecutor struct {
	client        *dagger.Client
	runner        *runnerpkg.Runner
	workflow      *gha.Workflow
	job           *gha.Job
	context       *gha.RunContext
	log           logger.Logger
	stepExecutors []StepExecutor
}

// NewJobExecutor creates a new job executor.
func NewJobExecutor(ctx context.Context, client *dagger.Client, workflow *gha.Workflow, job *gha.Job, context *gha.RunContext, log logger.Logger) (*JobExecutor, error) {
	// Create runner
	runner, err := runnerpkg.NewRunner(ctx, client)
	if err != nil {
		return nil, err
	}

	return &JobExecutor{
		client:        client,
		runner:        runner,
		workflow:      workflow,
		job:           job,
		context:       context,
		log:           log,
		stepExecutors: []StepExecutor{},
	}, nil
}

func (j *JobExecutor) Execute(ctx context.Context) error {
	if err := j.setup(ctx); err != nil {
		return err
	}

	for _, se := range j.stepExecutors {
		if err := se.pre(ctx, j.runner); err != nil {
			return err
		}
	}

	for _, se := range j.stepExecutors {
		if err := se.main(ctx, j.runner); err != nil {
			return err
		}
	}

	for _, se := range j.stepExecutors {
		if err := se.post(ctx, j.runner); err != nil {
			return err
		}
	}

	return nil
}

func (j *JobExecutor) setup(ctx context.Context) error {
	j.log.Info("Set up job")

	// TODO: this is a hack, we should find better way to do this
	j.runner.WithExec("mkdir", "-p", j.context.Github.Workspace)

	j.runner.WithEnvironment(j.context.ToEnv())
	j.runner.WithEnvironment(j.workflow.Environment)
	j.runner.WithEnvironment(j.job.Environment)

	for _, step := range j.job.Steps {
		j.runner.WithCustomAction(step.Uses)

		j.stepExecutors = append(j.stepExecutors, NewStepActionExecutor(step, j.log, j.context.ToEnv(), j.workflow.Environment, j.job.Environment))

		j.log.Info(fmt.Sprintf("Download action repository '%s'", step.Uses))
	}

	return nil
}
