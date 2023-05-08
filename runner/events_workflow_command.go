package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
)

var (
	_ event.Event[Context] = new(GithubWorkflowCommandEvent)
	_ event.Event[Context] = new(SetOutputEvent)
	_ event.Event[Context] = new(AddPathEvent)
	_ event.Event[Context] = new(AddMaskEvent)
	_ event.Event[Context] = new(AddMatcherEvent)
	_ event.Event[Context] = new(SaveStepStateEvent)
)

// GithubWorkflowCommandEvent is emitted when a command is emitted by a step. This is used to communicate with the
// runner environment from the step.
type GithubWorkflowCommandEvent struct {
	// raw is the raw command string. It is kept for debugging purposes. Command is already parsed by this point.
	Raw string

	// command is the parsed command
	Command *gha.Command

	// stepID is the ID of the step that emitted this command.
	StepID string
}

func (e GithubWorkflowCommandEvent) Handle(ctx context.Context, _ *Context, publisher event.Publisher[Context]) event.Result[Context] {
	var (
		stepID  = e.StepID
		command = e.Command
	)

	// Only handle commands that are make modifications to the environment. Logging commands are handled by the logger.
	switch command.Name {
	case "set-env":
		publisher.Publish(ctx, AddEnvEvent{Name: command.Parameters["name"], Value: command.Value})
	case "set-output":
		publisher.Publish(ctx, SetOutputEvent{StepID: stepID, Name: command.Parameters["name"], Value: command.Value})
	case "add-path":
		publisher.Publish(ctx, AddPathEvent{Path: command.Value})
	case "add-mask":
		publisher.Publish(ctx, AddMaskEvent{Value: command.Value})
	case "save-state":
		publisher.Publish(ctx, SaveStepStateEvent{StepID: stepID, Name: command.Parameters["name"], Value: command.Value})
	case "add-matcher":
		publisher.Publish(ctx, AddMatcherEvent{Matcher: command.Value})
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// SetOutputEvent is emitted when a step sets an output. It's triggered by the `set-output` command.
type SetOutputEvent struct {
	StepID string
	Name   string
	Value  string
}

func (e SetOutputEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	sr, ok := ec.stepResults[e.StepID]
	if !ok {
		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("step result %s not found", e.StepID)}
	}

	sr.Outputs[e.Name] = e.Value

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// AddPathEvent is emitted when a step adds a path to the PATH environment variable. It's triggered by the `add-path`
// command.
type AddPathEvent struct {
	Path string
}

func (e AddPathEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	path, err := ec.container.EnvVariable(ctx, "PATH")
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("failed to get PATH: %w", err)}
	}

	ec.container = ec.container.WithEnvVariable("PATH", fmt.Sprintf("%s:%s", path, e.Path))

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// AddMaskEvent is emitted when a step adds a mask to the environment. It's triggered by the `add-mask` command.
type AddMaskEvent struct {
	Value string
}

func (e AddMaskEvent) Handle(_ context.Context, _ *Context, _ event.Publisher[Context]) event.Result[Context] {
	fmt.Printf("add-mask: %s\n", e.Value)

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// AddMatcherEvent is emitted when a step adds a matcher to the environment. It's triggered by the `add-matcher` command.
type AddMatcherEvent struct {
	Matcher string
}

func (e AddMatcherEvent) Handle(_ context.Context, _ *Context, _ event.Publisher[Context]) event.Result[Context] {
	fmt.Printf("add-matcher: %s\n", e.Matcher)

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// SaveStepStateEvent is emitted when a step saves a state. It's triggered by the `save-state` command.
type SaveStepStateEvent struct {
	StepID string
	Name   string
	Value  string
}

func (e SaveStepStateEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	state, ok := ec.stepState[e.StepID]
	if !ok {
		state = make(map[string]string)
		ec.stepState[e.StepID] = state
	}

	state[e.Name] = e.Value

	return event.Result[Context]{Status: event.StatusSucceeded}
}
