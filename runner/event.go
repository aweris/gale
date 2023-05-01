package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"
	"github.com/google/uuid"

	"github.com/aweris/gale/gha"
)

type EventStatus string

const (
	EventStatusInProgress = "in_progress"
	EventStatusSucceeded  = "succeeded"
	EventStatusSkipped    = "skipped"
	EventStatusFailed     = "failed"
)

// Event represents a significant change or action that occurs within the runner.
type Event interface {
	handle(context.Context, *runner) EventResult
}

type EventResult struct {
	status   EventStatus
	err      error
	stdout   string
	children []*EventRecord
}

type EventRecord struct {
	Event
	EventResult

	ID int
}

func (r *runner) handle(ctx context.Context, event Event) *EventRecord {
	record := &EventRecord{
		ID:    int(r.counter.Add(1)),
		Event: event,
		EventResult: EventResult{
			status: EventStatusInProgress,
		},
	}

	r.events = append(r.events, record)

	record.EventResult = event.handle(ctx, r)

	return record
}

// newSuccessEvent creates a new success event without any stdout.
func newSuccessEvent() EventResult {
	return EventResult{status: EventStatusSucceeded}
}

// newSkippedEvent creates a new skipped event without any stdout.
func newSkippedEvent() EventResult {
	return EventResult{status: EventStatusSkipped}
}

// newErrorEvent creates a new error event without any stdout.
func newErrorEvent(err error) EventResult {
	return EventResult{status: EventStatusFailed, err: err}
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

func (e AddEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithEnvVariable(e.name, e.value)
	return newSuccessEvent()
}

// ReplaceEnvEvent replaces existing env value with the new one. Event assumes existing env and value validated
// during event creation. It's not validate again
type ReplaceEnvEvent struct {
	name     string
	oldValue string
	newValue string
}

func (e ReplaceEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithEnvVariable(e.name, e.newValue)
	return newSuccessEvent()
}

// RemoveEnvEvent removes an env value from runner container
type RemoveEnvEvent struct {
	name string
}

func (e RemoveEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithoutEnvVariable(e.name)
	return newSuccessEvent()
}

// Directory Events

var _ Event = new(WithDirectoryEvent)

// WithDirectoryEvent mounts a dagger.Directory to runner container with given path.
type WithDirectoryEvent struct {
	path string
	dir  *dagger.Directory
}

func (e WithDirectoryEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithDirectory(e.path, e.dir)
	return newSuccessEvent()
}

// Exec Events

var _ Event = new(WithExecEvent)

// WithExecEvent adds WithExec to runner container with given args.
type WithExecEvent struct {
	args []string
}

func (e WithExecEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithExec(e.args)
	return newSuccessEvent()
}

// Job Events

var (
	_ Event = new(SetupJobEvent)
)

// SetupJobEvent runs `setup job` step for the runner job.
type SetupJobEvent struct {
	// Intentionally left blank. It's not take any parameters
}

func (e SetupJobEvent) handle(ctx context.Context, runner *runner) EventResult {
	runner.log.Info("Set up job")

	// TODO: this is a hack, we should find better way to do this
	runner.WithExec("mkdir", "-p", runner.context.Github.Workspace)

	runner.WithEnvironment(runner.context.ToEnv())
	runner.WithEnvironment(runner.workflow.Environment)
	runner.WithEnvironment(runner.job.Environment)

	for _, step := range runner.job.Steps {
		runner.WithCustomAction(step.Uses)

		runner.log.Info(fmt.Sprintf("Download action repository '%s'", step.Uses))
	}

	return newSuccessEvent()
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

func (e WithActionEvent) handle(ctx context.Context, runner *runner) EventResult {
	action, err := gha.LoadActionFromSource(ctx, runner.client, e.source)
	if err != nil {
		return newErrorEvent(err)
	}

	path := fmt.Sprintf("/home/runner/_temp/%s", uuid.New())

	runner.actionsBySource[e.source] = action
	runner.actionPathsBySource[e.source] = path

	runner.container = runner.container.WithDirectory(path, action.Directory)

	return newSuccessEvent()
}

// ExecStepActionEvent executes step on runner
type ExecStepActionEvent struct {
	stage string
	step  *gha.Step
}

func (e ExecStepActionEvent) handle(ctx context.Context, runner *runner) EventResult {
	var (
		runs   = ""
		step   = e.step
		path   = runner.actionPathsBySource[step.Uses]
		action = runner.actionsBySource[step.Uses]
	)

	switch e.stage {
	case "pre":
		runs = action.Runs.Pre
	case "main":
		runs = action.Runs.Main
	case "post":
		runs = action.Runs.Post
	default:
		return newErrorEvent(fmt.Errorf("unknow stage %s for ExecActionEvent", e.stage))
	}

	if runs == "" {
		return newSkippedEvent()
	}

	// TODO: check if conditions

	runner.log.Info(fmt.Sprintf("%s Run %s", e.stage, step.Uses))

	var children []*EventRecord

	// Set up inputs and environment variables for step
	children = append(children, runner.WithEnvironment(step.Environment)...)
	children = append(children, runner.WithInputs(step.With)...)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	out, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, runs))

	// Clean up inputs and environment variables for next step

	children = append(children, runner.WithoutInputs(step.With)...)
	children = append(children, runner.WithoutEnvironment(step.Environment, runner.context.ToEnv(), runner.workflow.Environment, runner.job.Environment)...)

	return EventResult{
		status:   EventStatusSucceeded,
		err:      outErr,
		stdout:   out,
		children: children,
	}
}

// Runner Events

var (
	_ Event = new(BuildContainerEvent)
	_ Event = new(LoadContainerEvent)
)

// BuildContainerEvent builds a default runner container. This event will be called right before executing job if runner
// not exist yet.
type BuildContainerEvent struct {
	// Intentionally left blank
}

func (e BuildContainerEvent) handle(ctx context.Context, runner *runner) EventResult {
	container, err := NewBuilder(runner.client).build(ctx)
	if err != nil {
		return newErrorEvent(err)
	}

	runner.container = container

	return newSuccessEvent()
}

// LoadContainerEvent load container from given host path
type LoadContainerEvent struct {
	path string
}

func (e LoadContainerEvent) handle(ctx context.Context, runner *runner) EventResult {
	dir := filepath.Dir(e.path)
	base := filepath.Base(e.path)

	runner.container = runner.client.Container().Import(runner.client.Host().Directory(dir).File(base))

	return newSuccessEvent()
}
