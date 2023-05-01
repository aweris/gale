package runner

import (
	"context"
	"github.com/aweris/gale/gha"
)

// ExecAndCaptureOutput combines daggers WithExec and Stdout methods into one.
func (r *Runner) ExecAndCaptureOutput(ctx context.Context, cmd string, more ...string) (string, error) {
	return r.Container.WithExec(append([]string{cmd}, more...)).Stdout(ctx)
}

// ExecStepAction execute
func (r *Runner) ExecStepAction(ctx context.Context, stage string, step *gha.Step) {
	r.handle(ctx, ExecStepActionEvent{stage: stage, step: step})
}
