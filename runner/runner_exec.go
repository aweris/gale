package runner

import (
	"context"
)

// ExecAndCaptureOutput combines daggers WithExec and Stdout methods into one.
func (r *runner) ExecAndCaptureOutput(ctx context.Context, cmd string, more ...string) (string, error) {
	return r.container.WithExec(append([]string{cmd}, more...)).Stdout(ctx)
}
