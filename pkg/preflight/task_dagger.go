package preflight

import (
	"context"
	"strings"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
)

var _ Task = new(DaggerCheck)

// DaggerCheck is a preflight check that checks if dagger is working properly.
type DaggerCheck struct{}

func (d *DaggerCheck) Name() string {
	return NameDaggerCheck
}

func (d *DaggerCheck) Type() TaskType {
	return TaskTypeCheck
}

func (d *DaggerCheck) DependsOn() []string {
	return []string{}
}

func (d *DaggerCheck) Run(_ *Context, _ Options) Result {
	// check if dagger is initialized in global config
	client := config.Client()
	if client == nil {
		return Result{Status: Failed, Messages: []Message{{Level: Error, Content: "Global client is not initialized"}}}
	}

	var msgs []Message

	msgs = append(msgs, Message{Level: Info, Content: "Global client is initialized"})

	// initialize dagger context
	dctx := core.NewDaggerContextFromEnv()

	// this is not possible to happen here. Only added for ensuring that the code is working.
	if dctx == nil {
		msgs = append(msgs, Message{Level: Error, Content: "Dagger context is not initialized from environment"})

		return Result{Status: Failed, Messages: msgs}
	}

	// check if runner host or session is not provided then fallback to docker socket is exists
	if dctx.RunnerHost == "" || dctx.Session == "" {
		msgs = append(msgs, Message{Level: Info, Content: "Dagger runner host and/or session is not provided. Using docker socket instead."})

		socket := client.Host().UnixSocket(dctx.DockerSock)

		if _, err := socket.ID(context.Background()); err != nil {
			msgs = append(msgs, Message{Level: Error, Content: "Docker socket is not reachable"})

			return Result{Status: Failed, Messages: msgs}
		}

		msgs = append(msgs, Message{Level: Info, Content: "Docker socket is reachable"})
	}

	// check if we can execute a container with given dagger context exist in environment
	out, err := client.Container().
		From("alpine:latest").
		WithExec([]string{"echo", "Hello World"}).
		Stdout(context.Background())
	if err != nil {
		msgs = append(msgs, Message{Level: Error, Content: "Container execution is not successful"})

		return Result{Status: Failed, Messages: msgs}
	}

	if strings.TrimSpace(out) != "Hello World" {
		msgs = append(msgs, Message{Level: Error, Content: "Container output is not correct. Expected: Hello World, Actual: " + out})

		return Result{Status: Failed, Messages: msgs}
	}

	msgs = append(msgs, Message{Level: Info, Content: "Container execution is successful"})

	return Result{Status: Passed, Messages: msgs}
}
