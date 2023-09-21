package gctx

import (
	"os"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

// DockerContext represents the docker context for passing docker socket to the container.
type DockerContext struct {
	Sock string // Sock is the path to the docker socket.
}

// TODO: this is not enough. Most of users can't use gale because of this. We need to find a better way to handle this.

// LoadDaggerContext loads a new dagger context from environment variables
func (c *Context) LoadDaggerContext() error {
	// check if DOCKER_HOST should override the default docker socket location
	sock := "/var/run/docker.sock"

	if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
		sock = strings.TrimPrefix(dockerHost, "unix://")
	}

	c.Docker = DockerContext{Sock: sock}

	return nil
}

// helpers.WithContainerFuncHook interface to be loaded in the container.

var _ helpers.WithContainerFuncHook = new(DockerContext)

func (d *DockerContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.WithUnixSocket("/var/run/docker.sock", config.Client().Host().UnixSocket(d.Sock))
	}
}
