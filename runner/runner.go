package runner

import (
	"context"

	"github.com/aweris/gale/runner/state"
	"github.com/aweris/gale/runner/workflows"
)

var _ Runner = new(runner)

type Runner interface {
	RunWorkflow(ctx context.Context, name string) error // TBD -- result as well
}

// runner represents a GitHub Action runner powered by Dagger.
type runner struct {
	wh *workflows.Handler
}

// NewRunner creates a new Runner.
func NewRunner(base *state.BaseState) (Runner, error) {
	wh := workflows.NewHandler(state.NewWorkflowRunState(base))

	return &runner{wh: wh}, nil
}

func (r *runner) RunWorkflow(ctx context.Context, name string) error {
	return r.wh.RunWorkflow(ctx, name)
}
