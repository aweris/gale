package gale

import (
	"context"

	"dagger.io/dagger"
)

// TODO: move this file contents to gale.go after refactoring all the code

// With is an interface that can be implemented by any type that can be used to configure a dagger container.
type With interface {
	// WithContainerFunc is a function that can be used to configure the container. The signature of this function
	// matches the signature of the dagger.WithContainerFunc type. Intended to be used with the Load function.
	WithContainerFunc(container *dagger.Container) *dagger.Container
}

// Load is a helper function that can be used to load configuration into a dagger container in a more readable way.
func Load[T With](t T) dagger.WithContainerFunc {
	return t.WithContainerFunc
}

// fail returns a container that immediately fails with the given error. This useful for forcing a pipeline to fail
// inside chaining operations.
func fail(container *dagger.Container, err error) *dagger.Container {
	// fail the container with the given error
	container = container.WithExec([]string{"sh", "-c", "echo " + err.Error() + " && exit 1"})

	// forced evaluation of the pipeline to immediately fail
	container, _ = container.Sync(context.Background())

	return container
}
