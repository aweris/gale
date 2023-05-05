package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/event"
)

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
