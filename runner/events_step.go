package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
)

var _ event.Event[Context] = new(ExecStepEvent)

// ExecStepEvent orchestrates the execution of a step. It is responsible for setting up and cleaning up the step
// environment and publishing events to execute the step.
type ExecStepEvent struct {
	Stage gha.ActionStage
	Step  *gha.Step
}

func (e ExecStepEvent) Handle(ctx context.Context, ec *Context, publisher event.Publisher[Context]) event.Result[Context] {
	ec.log.Info(fmt.Sprintf("%s Run %s", e.Stage, e.Step.Name))

	// Set up state and environment variables for step

	if state, stateOK := ec.stepState[e.Step.ID]; stateOK {
		publisher.Publish(ctx, WithStepStateEvent{State: state})
	}

	if len(e.Step.Environment) > 0 {
		publisher.Publish(ctx, WithEnvironmentEvent{Env: e.Step.Environment})
	}

	// TODO: add check for the step type for shell, docker, etc. and publish the appropriate event. For now, we only support actions
	publisher.Publish(ctx, ExecStepActionEvent(e)) // convert to ExecStepActionEvent since events are identical

	// Clean up state and environment variables for next step

	if len(e.Step.Environment) > 0 {
		withoutEnv := WithoutEnvironmentEvent{
			Env:          e.Step.Environment,
			FallbackEnvs: []gha.Environment{ec.context.ToEnv(), ec.workflow.Environment, ec.job.Environment},
		}
		publisher.Publish(ctx, withoutEnv)
	}

	if state, stateOK := ec.stepState[e.Step.ID]; stateOK {
		publisher.Publish(ctx, WithoutStepStateEvent{State: state})
	}

	return event.Result[Context]{Status: event.StatusSucceeded}
}
