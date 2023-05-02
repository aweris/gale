package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

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
	Status   EventStatus
	Err      error
	Stdout   string
	Children []*EventRecord
}

type EventRecord struct {
	Event
	EventResult

	ID        int
	EventName string
	Timestamp time.Time
}

func (r *runner) handle(ctx context.Context, event Event) *EventRecord {
	record := &EventRecord{
		ID:        int(r.counter.Add(1)),
		EventName: reflect.TypeOf(event).Name(),
		Event:     event,
		EventResult: EventResult{
			Status: EventStatusInProgress,
		},
		Timestamp: time.Now(),
	}

	r.events = append(r.events, record)

	record.EventResult = event.handle(ctx, r)

	return record
}

// newSuccessEvent creates a new success event without any Stdout.
func newSuccessEvent() EventResult {
	return EventResult{Status: EventStatusSucceeded}
}

// newSkippedEvent creates a new skipped event without any Stdout.
func newSkippedEvent() EventResult {
	return EventResult{Status: EventStatusSkipped}
}

// newErrorEvent creates a new error event without any Stdout.
func newErrorEvent(err error) EventResult {
	return EventResult{Status: EventStatusFailed, Err: err}
}

// Environment Events

var (
	_ Event = new(WithEnvironmentEvent)
	_ Event = new(WithoutEnvironmentEvent)
	_ Event = new(AddEnvEvent)
	_ Event = new(ReplaceEnvEvent)
	_ Event = new(RemoveEnvEvent)
)

// WithEnvironmentEvent introduces new environment to runner container.
type WithEnvironmentEvent struct {
	Env gha.Environment
}

func (e WithEnvironmentEvent) handle(ctx context.Context, runner *runner) EventResult {
	var children []*EventRecord

	for k, v := range e.Env {
		if val, _ := runner.container.EnvVariable(ctx, k); val != "" {
			children = append(children, runner.handle(ctx, ReplaceEnvEvent{Name: k, OldValue: val, NewValue: v}))
		} else {
			children = append(children, runner.handle(ctx, AddEnvEvent{Name: k, Value: v}))
		}
	}

	return EventResult{Status: EventStatusSucceeded, Children: children}
}

// WithoutEnvironmentEvent removes given environment variables from the container. If a fallback environment is given,
// instead of removing the variable, it will be set to the value of the fallback environment.
//
// If multiple fallback environments are given, they will be merged in the order they are given. The last environment
// in the list will have the highest priority.
//
// This is useful for removing overridden environment variables without losing the original value.
type WithoutEnvironmentEvent struct {
	Env          gha.Environment
	FallbackEnvs []gha.Environment
}

func (e WithoutEnvironmentEvent) handle(ctx context.Context, runner *runner) EventResult {
	merged := gha.Environment{}

	for _, environment := range e.FallbackEnvs {
		// to merge the fallback environments with priority, we need to merge them in order.
		// the last environment in the list will have the highest priority.
		merged = merged.Merge(environment)
	}

	var children []*EventRecord

	for k, v := range e.Env {
		if _, ok := merged[k]; ok {
			children = append(children, runner.handle(ctx, ReplaceEnvEvent{Name: k, OldValue: v, NewValue: merged[k]}))
		} else {
			children = append(children, runner.handle(ctx, RemoveEnvEvent{Name: k}))
		}
	}

	return EventResult{Status: EventStatusSucceeded, Children: children}
}

// AddEnvEvent introduces new env variable to runner container.
type AddEnvEvent struct {
	Name  string
	Value string
}

func (e AddEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithEnvVariable(e.Name, e.Value)
	return newSuccessEvent()
}

// ReplaceEnvEvent replaces existing env Value with the new one. Event assumes existing env and Value validated
// during event creation. It's not validate again
type ReplaceEnvEvent struct {
	Name     string
	OldValue string
	NewValue string
}

func (e ReplaceEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithEnvVariable(e.Name, e.NewValue)
	return newSuccessEvent()
}

// RemoveEnvEvent removes an env Value from runner container
type RemoveEnvEvent struct {
	Name string
}

func (e RemoveEnvEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithoutEnvVariable(e.Name)
	return newSuccessEvent()
}

// Exec Events

var _ Event = new(WithExecEvent)

// WithExecEvent adds WithExec to runner container with given Args.
type WithExecEvent struct {
	Args []string
}

func (e WithExecEvent) handle(_ context.Context, runner *runner) EventResult {
	runner.container = runner.container.WithExec(e.Args)
	return newSuccessEvent()
}

// Job Events

var (
	_ Event = new(SetupJobEvent)
)

// SetupJobEvent runs `setup job` Step for the runner job.
type SetupJobEvent struct {
	// Intentionally left blank. It's not take any parameters
}

func (e SetupJobEvent) handle(ctx context.Context, runner *runner) EventResult {
	runner.log.Info("Set up job")

	var children []*EventRecord

	// TODO: this is a hack, we should find better way to do this
	children = append(children, runner.handle(ctx, WithExecEvent{Args: []string{"mkdir", "-p", runner.context.Github.Workspace}}))

	children = append(children, runner.handle(ctx, WithEnvironmentEvent{Env: runner.context.ToEnv()}))
	children = append(children, runner.handle(ctx, WithEnvironmentEvent{Env: runner.workflow.Environment}))
	children = append(children, runner.handle(ctx, WithEnvironmentEvent{Env: runner.job.Environment}))

	for _, step := range runner.job.Steps {
		children = append(children, runner.handle(ctx, WithActionEvent{Source: step.Uses}))

		runner.log.Info(fmt.Sprintf("Download action repository '%s'", step.Uses))
	}

	return EventResult{Status: EventStatusSucceeded, Children: children}
}

// Action Events

var (
	_ Event = new(WithStepInputsEvent)
	_ Event = new(WithoutStepInputsEvent)
	_ Event = new(WithActionEvent)
	_ Event = new(ExecStepActionEvent)
)

// WithStepInputsEvent transform given input name as INPUT_<NAME> and add it to the container as environment variable.
type WithStepInputsEvent struct {
	Inputs map[string]string
}

func (e WithStepInputsEvent) handle(ctx context.Context, runner *runner) EventResult {
	for k, v := range e.Inputs {
		// TODO: This is a hack to get around the fact that we can't set the GITHUB_TOKEN as an input. Remove this
		// once we have a better solution.
		if strings.TrimSpace(v) == "${{ secrets.GITHUB_TOKEN }}" {
			v = os.Getenv("GITHUB_TOKEN")
		}

		runner.container = runner.container.WithEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)), v)
	}

	return newSuccessEvent()
}

// WithoutStepInputsEvent removes the given inputs from the container.
type WithoutStepInputsEvent struct {
	Inputs map[string]string
}

func (e WithoutStepInputsEvent) handle(ctx context.Context, runner *runner) EventResult {
	for k := range e.Inputs {
		runner.container = runner.container.WithoutEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)))
	}

	return newSuccessEvent()
}

// WithActionEvent fetches github action code from given Source and mount as a directory in a runner container.
type WithActionEvent struct {
	Source string
}

func (e WithActionEvent) handle(ctx context.Context, runner *runner) EventResult {
	action, err := gha.LoadActionFromSource(ctx, runner.client, e.Source)
	if err != nil {
		return newErrorEvent(err)
	}

	path := fmt.Sprintf("/home/runner/_temp/%s", uuid.New())

	runner.actionsBySource[e.Source] = action
	runner.actionPathsBySource[e.Source] = path

	runner.container = runner.container.WithDirectory(path, action.Directory)

	return newSuccessEvent()
}

// ExecStepActionEvent executes Step on runner
type ExecStepActionEvent struct {
	Stage string
	Step  *gha.Step
}

func (e ExecStepActionEvent) handle(ctx context.Context, runner *runner) EventResult {
	var (
		runs   = ""
		step   = e.Step
		path   = runner.actionPathsBySource[step.Uses]
		action = runner.actionsBySource[step.Uses]
	)

	switch e.Stage {
	case "pre":
		runs = action.Runs.Pre
	case "main":
		runs = action.Runs.Main
	case "post":
		runs = action.Runs.Post
	default:
		return newErrorEvent(fmt.Errorf("unknow stage %s for ExecActionEvent", e.Stage))
	}

	if runs == "" {
		return newSkippedEvent()
	}

	// TODO: check if conditions

	runner.log.Info(fmt.Sprintf("%s Run %s", e.Stage, step.Uses))

	var children []*EventRecord

	// Set up inputs and environment variables for step
	children = append(children, runner.handle(ctx, WithEnvironmentEvent{Env: step.Environment}))
	children = append(children, runner.handle(ctx, WithStepInputsEvent{Inputs: step.With}))

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	out, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, runs))

	// Clean up inputs and environment variables for next step

	children = append(children, runner.handle(ctx, WithoutStepInputsEvent{Inputs: step.With}))

	withoutEnv := WithoutEnvironmentEvent{
		Env:          step.Environment,
		FallbackEnvs: []gha.Environment{runner.context.ToEnv(), runner.workflow.Environment, runner.job.Environment},
	}
	children = append(children, runner.handle(ctx, withoutEnv))

	return EventResult{
		Status:   EventStatusSucceeded,
		Err:      outErr,
		Stdout:   out,
		Children: children,
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
	Path string
}

func (e LoadContainerEvent) handle(_ context.Context, runner *runner) EventResult {
	dir := filepath.Dir(e.Path)
	base := filepath.Base(e.Path)

	runner.container = runner.client.Container().Import(runner.client.Host().Directory(dir).File(base))

	return newSuccessEvent()
}
