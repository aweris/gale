package runner

import (
	"context"

	"dagger.io/dagger"
)

// Runner represents a GitHub Action runner powered by Dagger.
type Runner struct {
	// Container is the Dagger container that the runner is running in.
	Container *dagger.Container
}

// NewRunner creates a new Runner.
func NewRunner(ctx context.Context, client *dagger.Client) (*Runner, error) {
	runner := NewBuilder(client).From("docker.io/library/ubuntu:22.04").Build(ctx)

	// run as non-root user runner like the original runner image
	runner.Container = runner.Container.WithUser("runner")

	// provision the container before returning in order to fail early if there are any issues
	_, err := runner.Container.ExitCode(ctx)
	if err != nil {
		return nil, err
	}

	return runner, nil
}
