package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
)

// Environment Events

var (
	_ event.Event[Context] = new(WithEnvironmentEvent)
	_ event.Event[Context] = new(WithoutEnvironmentEvent)
	_ event.Event[Context] = new(AddEnvEvent)
	_ event.Event[Context] = new(ReplaceEnvEvent)
	_ event.Event[Context] = new(RemoveEnvEvent)
)

// WithEnvironmentEvent introduces new environment to runner container.
type WithEnvironmentEvent struct {
	Env gha.Environment
}

func (e WithEnvironmentEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
	for k, v := range e.Env {
		if val, _ := ec.container.EnvVariable(ctx, k); val != "" {
			publisher.Publish(ctx, ReplaceEnvEvent{Name: k, OldValue: val, NewValue: v})
		} else {
			publisher.Publish(ctx, AddEnvEvent{Name: k, Value: v})
		}
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
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

func (e WithoutEnvironmentEvent) Handle(ctx context.Context, _ *Context, publisher event.Publisher[Context]) event.Result[Context] {
	merged := gha.Environment{}

	for _, environment := range e.FallbackEnvs {
		// to merge the fallback environments with priority, we need to merge them in order.
		// the last environment in the list will have the highest priority.
		merged = merged.Merge(environment)
	}

	for k, v := range e.Env {
		if _, ok := merged[k]; ok {
			publisher.Publish(ctx, ReplaceEnvEvent{Name: k, OldValue: v, NewValue: merged[k]})
		} else {
			publisher.Publish(ctx, RemoveEnvEvent{Name: k})
		}
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// AddEnvEvent introduces new env variable to runner container.
type AddEnvEvent struct {
	Name  string
	Value string
}

func (e AddEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithEnvVariable(e.Name, e.Value)
	return event.Result[Context]{Status: event.StatusSucceeded}
}

// ReplaceEnvEvent replaces existing env Value with the new one. Event assumes existing env and Value validated
// during event creation. It's not validate again
type ReplaceEnvEvent struct {
	Name     string
	OldValue string
	NewValue string
}

func (e ReplaceEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithEnvVariable(e.Name, e.NewValue)
	return event.Result[Context]{Status: event.StatusSucceeded}
}

// RemoveEnvEvent removes an env Value from runner container
type RemoveEnvEvent struct {
	Name string
}

func (e RemoveEnvEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithoutEnvVariable(e.Name)
	return event.Result[Context]{Status: event.StatusSucceeded}
}

// Exec Events

var _ event.Event[Context] = new(WithExecEvent)

// WithExecEvent adds WithExec to runner container with given Args. If Execute is true, it will execute the command
// immediately after adding it to the container. Otherwise, it will be added to the container but not executed.
type WithExecEvent struct {
	Args    []string
	Execute bool
}

func (e WithExecEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithExec(e.Args)

	if !e.Execute {
		return event.Result[Context]{Status: event.StatusSucceeded}
	}

	out, err := ec.container.Stdout(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err, Stdout: out}
	}

	return event.Result[Context]{Status: event.StatusSucceeded, Stdout: out}
}

// Job Events

var (
	_ event.Event[Context] = new(SetupJobEvent)
)

// SetupJobEvent runs `setup job` Step for the runner job.
type SetupJobEvent struct {
	// Intentionally left blank. It's not take any parameters
}

func (e SetupJobEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
	ec.log.Info("Set up job")

	// TODO: this is a hack, we should find better way to do this
	publisher.Publish(ctx, WithExecEvent{Args: []string{"mkdir", "-p", ec.context.Github.Workspace}})

	publisher.Publish(ctx, WithEnvironmentEvent{Env: ec.context.ToEnv()})
	publisher.Publish(ctx, WithEnvironmentEvent{Env: ec.workflow.Environment})
	publisher.Publish(ctx, WithEnvironmentEvent{Env: ec.job.Environment})

	for idx, step := range ec.job.Steps {
		// using index as step ID if it's not set as fallback
		if step.ID == "" {
			step.ID = fmt.Sprintf("%d", idx)
		}

		publisher.Publish(ctx, WithActionEvent{Source: step.Uses})

		ec.log.Info(fmt.Sprintf("Download action repository '%s'", step.Uses))
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}

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

// Runner Events

var (
	_ event.Event[Context] = new(BuildContainerEvent)
	_ event.Event[Context] = new(LoadContainerEvent)
)

// BuildContainerEvent builds a default runner container. This event will be called right before executing job if runner
// not exist yet.
type BuildContainerEvent struct {
	// Intentionally left blank
}

func (e BuildContainerEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	container, err := NewBuilder(ec.client).build(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err}
	}

	ec.container = container

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// LoadContainerEvent load container from given host path
type LoadContainerEvent struct {
	Path string
}

func (e LoadContainerEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	dir := filepath.Dir(e.Path)
	base := filepath.Base(e.Path)

	ec.container = ec.client.Container().Import(ec.client.Host().Directory(dir).File(base))

	return event.Result[Context]{Status: event.StatusSucceeded}
}
