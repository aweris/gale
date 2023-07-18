package gale

import (
	"context"
	"fmt"
	"os"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/dagger/services"
	"github.com/aweris/gale/internal/dagger/tools"
	"github.com/aweris/gale/internal/model"
)

var _ With = new(Env)

// Env represents gale execution environment
type Env struct {
	client *dagger.Client

	// services
	artifacts *services.ArtifactService
}

func (g *Gale) Env() *Env {
	return &Env{client: g.client}
}

func (e *Env) WithContainerFunc(container *dagger.Container) *dagger.Container {
	// dagger context
	dc := model.NewDaggerContextFromEnv()

	for k, v := range dc.ToEnv() {
		container = container.WithEnvVariable(k, v)
	}

	// as a fallback, we're loading docker socket from the host.
	if dc.RunnerHost == "" || dc.Session == "" {
		// check if DOCKER_HOST should override the default docker socket location
		hostDockerSocket := "/var/run/docker.sock"

		if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
			hostDockerSocket = strings.TrimPrefix(dockerHost, "unix://")
		}

		container = container.WithUnixSocket("/var/run/docker.sock", e.client.Host().UnixSocket(hostDockerSocket))
	}

	// tools

	ghx, err := tools.Ghx(context.Background(), e.client)
	if err != nil {
		fail(container, fmt.Errorf("error getting ghx: %w", err))
	}

	container = container.WithFile("/usr/local/bin/ghx", ghx)

	// services

	artifactService := services.NewArtifactService(e.client)

	container = container.With(artifactService.ServiceBinding)

	e.artifacts = artifactService // save for later access

	return container
}
