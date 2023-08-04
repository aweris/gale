package helpers

import (
	"context"

	"dagger.io/dagger"
)

// FailPipeline returns a container that immediately fails with the given error. This useful for forcing a pipeline to
// fail inside chaining operations.
func FailPipeline(container *dagger.Container, err error) *dagger.Container {
	// fail the container with the given error
	container = container.WithExec([]string{"sh", "-c", "echo " + err.Error() + " && exit 1"})

	// forced evaluation of the pipeline to immediately fail
	container, _ = container.Sync(context.Background())

	return container
}

// WithContainerFuncHook is an interface that ensures that implementers have a WithContainerFunc method.
type WithContainerFuncHook interface {
	// WithContainerFunc returns a dagger function that allows the user to modify the container.
	WithContainerFunc() dagger.WithContainerFunc
}
