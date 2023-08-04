package core

import (
	"os"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

var _ helpers.WithContainerFuncHook = new(DaggerContext)

// DaggerContext represents the dagger engine connection information should be passed to the container
type DaggerContext struct {
	RunnerHost string // RunnerHost where the dagger engine is running
	Session    string // Session is the dagger session information
	DockerSock string // DockerSock is the path to the docker socket. This is used to as a fallback when the dagger host or session is not provided.
}

// NewDaggerContextFromEnv creates a new dagger context from environment variables
func NewDaggerContextFromEnv() *DaggerContext {
	// check if DOCKER_HOST should override the default docker socket location
	hostDockerSocket := "/var/run/docker.sock"

	if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
		hostDockerSocket = strings.TrimPrefix(dockerHost, "unix://")
	}

	return &DaggerContext{
		RunnerHost: os.Getenv("_EXPERIMENTAL_DAGGER_RUNNER_HOST"),
		Session:    os.Getenv("DAGGER_SESSION"),
		DockerSock: hostDockerSocket,
	}
}

func (d *DaggerContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		if d.RunnerHost != "" {
			container = container.WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", d.RunnerHost)
		}

		if d.Session != "" {
			container = container.WithEnvVariable("DAGGER_SESSION", d.Session)
		}

		// as a fallback, we're loading docker socket from the host.
		if d.RunnerHost == "" || d.Session == "" {
			container = container.WithUnixSocket("/var/run/docker.sock", config.Client().Host().UnixSocket(d.DockerSock))
		}

		return container
	}
}
