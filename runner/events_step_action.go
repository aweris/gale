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

var (
	_ event.Event[Context] = new(WithStepInputsEvent)
	_ event.Event[Context] = new(WithoutStepInputsEvent)
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
	Stage gha.ActionStage
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
	case gha.ActionStagePre:
		runs = action.Runs.Pre
	case gha.ActionStageMain:
		runs = action.Runs.Main
		ec.stepResults[step.ID] = &gha.StepResult{
			Outputs:    make(map[string]string),
			Conclusion: gha.StepStatusSuccess,
			Outcome:    gha.StepStatusSuccess,
		}
	case gha.ActionStagePost:
		runs = action.Runs.Post
	default:
		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("unknown stage %s", e.Stage)}
	}

	// if runs is empty for pre or post, this is a no-op step
	if runs == "" && e.Stage != gha.ActionStageMain {
		return event.Result[Context]{Status: event.StatusSkipped}
	}

	if runs == "" && e.Stage == gha.ActionStageMain {
		// update step result
		ec.stepResults[step.ID].Conclusion = gha.StepStatusFailure
		ec.stepResults[step.ID].Outcome = gha.StepStatusFailure

		return event.Result[Context]{Status: event.StatusFailed, Err: fmt.Errorf("no runs for step %s", step.ID)}
	}

	// TODO: check if conditions

	ec.log.Info(fmt.Sprintf("%s Run %s", e.Stage, step.Uses))

	// Set up inputs variables for step

	if len(step.With) > 0 {
		publisher.Publish(ctx, WithStepInputsEvent{Inputs: step.With})
	}

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well

	withExec := WithExecEvent{
		Args:    []string{"node", fmt.Sprintf("%s/%s", path, runs)},
		Execute: true,
		Strace:  true,
	}

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

	// Clean up state, inputs and environment variables for next step

	if len(step.With) > 0 {
		publisher.Publish(ctx, WithoutStepInputsEvent{Inputs: step.With})
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}
