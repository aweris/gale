package runner

import (
	"context"
	"dagger.io/dagger"
)

// Event represents a significant change or action that occurs within the runner.
type Event interface {
	handle(context.Context, *Runner)
}

func (r *Runner) handle(ctx context.Context, event Event) {
	r.events = append(r.events, event)

	event.handle(ctx, r)
}

// Environment Events

var (
	_ Event = new(AddEnvEvent)
	_ Event = new(ReplaceEnvEvent)
	_ Event = new(RemoveEnvEvent)
)

// AddEnvEvent introduces new env variable to runner container.
type AddEnvEvent struct {
	name  string
	value string
}

func (e AddEnvEvent) handle(_ context.Context, runner *Runner) {
	runner.Container = runner.Container.WithEnvVariable(e.name, e.value)
}

// ReplaceEnvEvent replaces existing env value with the new one. Event assumes existing env and value validated
// during event creation. It's not validate again
type ReplaceEnvEvent struct {
	name     string
	oldValue string
	newValue string
}

func (e ReplaceEnvEvent) handle(_ context.Context, runner *Runner) {
	runner.Container = runner.Container.WithEnvVariable(e.name, e.newValue)
}

// RemoveEnvEvent removes an env value from runner container
type RemoveEnvEvent struct {
	name string
}

func (e RemoveEnvEvent) handle(_ context.Context, runner *Runner) {
	runner.Container = runner.Container.WithoutEnvVariable(e.name)
}

// Directory Events

var _ Event = new(WithDirectoryEvent)

// WithDirectoryEvent mounts a dagger.Directory to runner container with given path.
type WithDirectoryEvent struct {
	path string
	dir  *dagger.Directory
}

func (e WithDirectoryEvent) handle(_ context.Context, runner *Runner) {
	runner.Container = runner.Container.WithDirectory(e.path, e.dir)
}

// Exec Events

var _ Event = new(WithExecEvent)

// WithExecEvent adds WithExec to runner container with given args.
type WithExecEvent struct {
	args []string
}

func (e WithExecEvent) handle(_ context.Context, runner *Runner) {
	runner.Container = runner.Container.WithExec(e.args)
}
