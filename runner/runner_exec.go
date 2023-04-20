package runner

import "context"

// ExecAndCaptureOutput combines daggers WithExec and Stdout methods into one.
func (r *Runner) ExecAndCaptureOutput(ctx context.Context, cmd string, more ...string) (string, error) {
	return r.Container.WithExec(append([]string{cmd}, more...)).Stdout(ctx)
}
