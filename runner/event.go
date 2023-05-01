package runner

import (
	"context"
	"fmt"

	"dagger.io/dagger"
	"github.com/google/uuid"

	"github.com/aweris/gale/gha"
)

// Event represents a significant change or action that occurs within the runner.
type Event interface {
	handle(context.Context, *Runner) error
}

func (r *Runner) handle(ctx context.Context, event Event) {
	r.events = append(r.events, event)

	if err := event.handle(ctx, r); err != nil {
		// TODO: handle errors
		fmt.Printf("Event failed %v", err)
	}
}

// Environment Events

var (
	_ Event = new(AddEnvEvent)
	_ Event = new(ReplaceEnvEvent)
	_ Event = new(RemoveEnvEvent)
)

// AddEnvEvent introduces new env variable to runner container.
type AddEnvEvent struct {
	name  string
	value string
}

func (e AddEnvEvent) handle(_ context.Context, runner *Runner) error {
	runner.Container = runner.Container.WithEnvVariable(e.name, e.value)
	return nil
}

// ReplaceEnvEvent replaces existing env value with the new one. Event assumes existing env and value validated
// during event creation. It's not validate again
type ReplaceEnvEvent struct {
	name     string
	oldValue string
	newValue string
}

func (e ReplaceEnvEvent) handle(_ context.Context, runner *Runner) error {
	runner.Container = runner.Container.WithEnvVariable(e.name, e.newValue)
	return nil
}

// RemoveEnvEvent removes an env value from runner container
type RemoveEnvEvent struct {
	name string
}

func (e RemoveEnvEvent) handle(_ context.Context, runner *Runner) error {
	runner.Container = runner.Container.WithoutEnvVariable(e.name)
	return nil
}

// Directory Events

var _ Event = new(WithDirectoryEvent)

// WithDirectoryEvent mounts a dagger.Directory to runner container with given path.
type WithDirectoryEvent struct {
	path string
	dir  *dagger.Directory
}

func (e WithDirectoryEvent) handle(_ context.Context, runner *Runner) error {
	runner.Container = runner.Container.WithDirectory(e.path, e.dir)
	return nil
}

// Exec Events

var _ Event = new(WithExecEvent)

// WithExecEvent adds WithExec to runner container with given args.
type WithExecEvent struct {
	args []string
}

func (e WithExecEvent) handle(_ context.Context, runner *Runner) error {
	runner.Container = runner.Container.WithExec(e.args)
	return nil
}

// Action Events

var (
	_ Event = new(WithActionEvent)
	_ Event = new(ExecStepActionEvent)
)

// WithActionEvent fetches github action code from given source and mount as a directory in a runner container.
type WithActionEvent struct {
	source string
}

func (e WithActionEvent) handle(ctx context.Context, runner *Runner) error {
	action, err := gha.LoadActionFromSource(ctx, runner.Client, e.source)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/home/runner/_temp/%s", uuid.New())

	runner.ActionsBySource[e.source] = action
	runner.ActionPathsBySource[e.source] = path

	runner.Container = runner.Container.WithDirectory(path, action.Directory)

	return nil
}

// ExecStepActionEvent executes step on runner
type ExecStepActionEvent struct {
	stage string
	step  *gha.Step
}

func (e ExecStepActionEvent) handle(ctx context.Context, runner *Runner) error {
	var (
		runs   = ""
		step   = e.step
		path   = runner.ActionPathsBySource[step.Uses]
		action = runner.ActionsBySource[step.Uses]
	)

	switch e.stage {
	case "pre":
		runs = action.Runs.Pre
	case "main":
		runs = action.Runs.Main
	case "post":
		runs = action.Runs.Post
	default:
		return fmt.Errorf("unknow stage %s for ExecActionEvent", e.stage)
	}

	if runs == "" {
		return nil
	}

	// TODO: check if conditions

	runner.log.Info(fmt.Sprintf("Pre Run %s", step.Uses))

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	_, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, runs))
	if outErr != nil {
		return outErr
	}

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, runner.context.ToEnv(), runner.workflow.Environment, runner.job.Environment)

	return nil
}
