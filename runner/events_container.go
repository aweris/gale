package runner

import (
	"context"
	"path/filepath"

	"github.com/aweris/gale/internal/event"
)

var (
	_ event.Event[Context] = new(BuildContainerEvent)
	_ event.Event[Context] = new(LoadContainerEvent)
	_ event.Event[Context] = new(WithExecEvent)
)

// BuildContainerEvent builds a default runner container. This event will be called right before executing job if runner
// not exist yet.
type BuildContainerEvent struct {
	// Intentionally left blank
}

func (e BuildContainerEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	container, err := NewBuilder(ec.client).build(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err}
	}

	ec.container = container

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// LoadContainerEvent load container from given host path. This event assumes that image export exists in given path.
type LoadContainerEvent struct {
	Path string
}

func (e LoadContainerEvent) Handle(_ context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	dir := filepath.Dir(e.Path)
	base := filepath.Base(e.Path)

	ec.container = ec.client.Container().Import(ec.client.Host().Directory(dir).File(base))

	return event.Result[Context]{Status: event.StatusSucceeded}
}

// WithExecEvent adds WithExec to runner container with given Args. If Execute is true, it will execute the command
// immediately after adding it to the container. Otherwise, it will be added to the container but not executed.
type WithExecEvent struct {
	Args    []string
	Execute bool
}

func (e WithExecEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	ec.container = ec.container.WithExec(e.Args)

	if !e.Execute {
		return event.Result[Context]{Status: event.StatusSucceeded}
	}

	out, err := ec.container.Stdout(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err, Stdout: out}
	}

	return event.Result[Context]{Status: event.StatusSucceeded, Stdout: out}
}
