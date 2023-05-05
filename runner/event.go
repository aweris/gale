package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
)

// Action Events

var (
	_ event.Event[Context] = new(WithStepInputsEvent)
	_ event.Event[Context] = new(WithoutStepInputsEvent)
	_ event.Event[Context] = new(SaveStepStateEvent)
	_ event.Event[Context] = new(WithStepStateEvent)
	_ event.Event[Context] = new(WithoutStepStateEvent)
	_ event.Event[Context] = new(WithActionEvent)
	_ event.Event[Context] = new(ExecStepActionEvent)
)

// WithStepInputsEvent transform given input name as INPUT_<NAME> and add it to the container as environment variable.
type WithStepInputsEvent struct {
	Inputs map[string]string
}

func (e WithStepInputsEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	for k, v := range e.Inputs {
		// TODO: This is a hack to get around the fact that we can't set the GITHUB_TOKEN as an input. Remove this
		// once we have a better solution.
		if strings.TrimSpace(v) == "${{ secrets.GITHUB_TOKEN }}" {
			v = os.Getenv("GITHUB_TOKEN")
		}

		ec.container = ec.container.WithEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)), v)
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// WithoutStepInputsEvent removes the given inputs from the container.
type WithoutStepInputsEvent struct {
	Inputs map[string]string
}

func (e WithoutStepInputsEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	for k := range e.Inputs {
		ec.container = ec.container.WithoutEnvVariable(fmt.Sprintf("INPUT_%s", strings.ToUpper(k)))
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

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

// WithStepStateEvent adds given state to the container as environment variable.
type WithStepStateEvent struct {
	State map[string]string
}

func (e WithStepStateEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	for k, v := range e.State {
		ec.container = ec.container.WithEnvVariable(fmt.Sprintf("STATE_%s", k), v)
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// WithoutStepStateEvent removes given state from the container.
type WithoutStepStateEvent struct {
	State map[string]string
}

func (e WithoutStepStateEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	for k := range e.State {
		ec.container = ec.container.WithoutEnvVariable(fmt.Sprintf("STATE_%s", k))
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// WithActionEvent fetches github action code from given Source and mount as a directory in a runner container.
type WithActionEvent struct {
	Source string
}

func (e WithActionEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	action, err := gha.LoadActionFromSource(ctx, ec.client, e.Source)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err}
	}

	path := fmt.Sprintf("/home/runner/_temp/%s", uuid.New())

	ec.actionsBySource[e.Source] = action
	ec.actionPathsBySource[e.Source] = path

	ec.container = ec.container.WithDirectory(path, action.Directory)

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// ExecStepActionEvent executes Step on runner
type ExecStepActionEvent struct {
	Stage string
	Step  *gha.Step
}

func (e ExecStepActionEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
	var (
		runs   string
		step   = e.Step
		path   = ec.actionPathsBySource[step.Uses]
		action = ec.actionsBySource[step.Uses]
	)

	switch e.Stage {
	case "pre":
		runs = action.Runs.Pre
	case "main":
		runs = action.Runs.Main
		ec.stepResults[step.ID] = &gha.StepResult{
			Outputs:    make(map[string]string),
			Conclusion: gha.StepStatusSuccess,
			Outcome:    gha.StepStatusSuccess,
		}
	case "post":
		runs = action.Runs.Post
	default:
		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("unknown stage %s", e.Stage)}
	}

	// if runs is empty for pre or post, this is a no-op step
	if runs == "" && e.Stage != "main" {
		return event.Result[Context]{Status: event.StatusSkipped}
	}

	if runs == "" && e.Stage == "main" {
		// update step result
		ec.stepResults[step.ID].Conclusion = gha.StepStatusFailure
		ec.stepResults[step.ID].Outcome = gha.StepStatusFailure

		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("no runs for step %s", step.ID)}
	}

	// TODO: check if conditions

	ec.log.Info(fmt.Sprintf("%s Run %s", e.Stage, step.Uses))

	// Set up inputs and environment variables for step
	state, stateOK := ec.stepState[step.ID]
	if stateOK {
		publisher.Publish(ctx, WithStepStateEvent{State: state})
	}

	publisher.Publish(ctx, WithEnvironmentEvent{Env: step.Environment})
	publisher.Publish(ctx, WithStepInputsEvent{Inputs: step.With})

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well

	withExec := WithExecEvent{Args: []string{"node", fmt.Sprintf("%s/%s", path, runs)}, Execute: true}

	record := publisher.Publish(ctx, withExec)

	if e.Stage == "main" && record.Status == event.StatusFailed {
		// TODO: check if step continue-on-error
		// update step result
		ec.stepResults[step.ID].Conclusion = gha.StepStatusFailure
		ec.stepResults[step.ID].Outcome = gha.StepStatusFailure
	}

	scanner := bufio.NewScanner(strings.NewReader(record.Stdout))

	// Loop through each line and process it
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			continue
		}

		isCommand, command := gha.ParseCommand(line)
		if !isCommand {
			continue
		}

		publisher.Publish(ctx, GithubWorkflowCommandEvent{Raw: line, Command: command, StepID: step.ID})
	}

	// Clean up inputs and environment variables for next step

	publisher.Publish(ctx, WithoutStepInputsEvent{Inputs: step.With})

	withoutEnv := WithoutEnvironmentEvent{
		Env:          step.Environment,
		FallbackEnvs: []gha.Environment{ec.context.ToEnv(), ec.workflow.Environment, ec.job.Environment},
	}
	publisher.Publish(ctx, withoutEnv)

	if stateOK {
		publisher.Publish(ctx, WithoutStepStateEvent{State: state})
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// GitHub Actions Events
var (
	_ event.Event[Context] = new(GithubWorkflowCommandEvent)
	_ event.Event[Context] = new(SetOutputEvent)
	_ event.Event[Context] = new(AddPathEvent)
	_ event.Event[Context] = new(AddMaskEvent)
	_ event.Event[Context] = new(AddMatcherEvent)
)

type GithubWorkflowCommandEvent struct {
	// raw is the raw command string. It is kept for debugging purposes. Command is already parsed by this point.
	Raw string

	// command is the parsed command
	Command *gha.Command

	// stepID is the ID of the step that emitted this command.
	StepID string
}

func (e GithubWorkflowCommandEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
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

type AddMaskEvent struct {
	Value string
}

func (e AddMaskEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	fmt.Printf("add-mask: %s\n", e.Value)

	return event.Result[Context]{Status: event.StatusSucceeded}
}

type AddMatcherEvent struct {
	Matcher string
}

func (e AddMatcherEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	fmt.Printf("add-matcher: %s\n", e.Matcher)

	return event.Result[Context]{Status: event.StatusSucceeded}
}
