package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/aweris/gale/builder"
	"github.com/aweris/gale/github/cli"
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
	repo, err := cli.CurrentRepository(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err}
	}

	container, err := builder.NewBuilder(ec.client, repo).Build(ctx)
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
	// Args is the command to be executed.
	Args []string

	// Execute is a flag to indicate whether the command should be executed immediately after adding it to the container.
	Execute bool

	// Strace is a flag to indicate whether the command should be executed with system trace. This is only applicable
	// if Execute is true.
	Strace bool
}

func (e WithExecEvent) Handle(ctx context.Context, ec *Context, _ event.Publisher[Context]) event.Result[Context] {
	var (
		args          = e.Args
		straceLogPath = fmt.Sprintf("/tmp/strace-%s.log", uuid.New())
	)

	if e.Execute && e.Strace {
		args = append([]string{"strace", "-o", straceLogPath}, args...)
	}

	ec.container = ec.container.WithExec(args)

	if !e.Execute {
		return event.Result[Context]{Status: event.StatusSucceeded}
	}

	out, err := ec.container.Stdout(ctx)
	if err != nil {
		return event.Result[Context]{Status: event.StatusFailed, Err: err, Stdout: out}
	}

	strace := ""

	if e.Strace {
		contents, err := ec.container.File(straceLogPath).Contents(ctx)
		if err != nil {
			return event.Result[Context]{
				Status: event.StatusFailed,
				Err:    fmt.Errorf("failed to read strace log: %w", err),
				Stdout: out,
				Strace: contents,
			}
		}

		strace = contents
	}

	return event.Result[Context]{Status: event.StatusSucceeded, Stdout: out, Strace: strace}
}
