package images

import "dagger.io/dagger"

// RunnerBase returns a container with the base image for the runner.
func RunnerBase(client *dagger.Client) *dagger.Container {
	return client.Container().
		Pipeline("Images").Pipeline("Runner Base").
		From("ghcr.io/catthehacker/ubuntu:act-22.04")
}
