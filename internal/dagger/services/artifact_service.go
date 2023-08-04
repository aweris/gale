package services

import (
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
)

var _ helpers.WithContainerFuncHook = new(ArtifactService)

// ArtifactService is the dagger service definitions for the services/artifact directory.
type ArtifactService struct {
	client    *dagger.Client
	container *dagger.Container
	artifacts *dagger.CacheVolume
	alias     string
	port      string
}

// NewArtifactService creates a new artifact service.
func NewArtifactService() *ArtifactService {
	v := version.GetVersion()

	tag := v.GitVersion

	container := config.Client().Container().From("ghcr.io/aweris/gale/services/artifact:" + tag)

	// port configuration
	container = container.WithEnvVariable("PORT", "8080").WithExposedPort(8080)

	// stateful data configuration

	cache := config.Client().CacheVolume("gale-artifact-service")
	container = container.WithMountedCache("/artifacts", cache).WithEnvVariable("ARTIFACTS_DIR", "/artifacts")

	return &ArtifactService{
		client:    config.Client(),
		container: container,
		artifacts: cache,
		alias:     "artifacts",
		port:      "8080",
	}
}

// Container returns the container of the artifact service.
func (a *ArtifactService) Container() *dagger.Container {
	return a.container
}

func (a *ArtifactService) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithServiceBinding(a.alias, a.container).
			WithEnvVariable("ACTIONS_RUNTIME_URL", fmt.Sprintf("http://%s:%s/", a.alias, a.port)).
			WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token") // dummy token, not used by service
	}
}

// Artifacts returns a artifact directory for the given run ID.
func (a *ArtifactService) Artifacts(runID string) *dagger.Directory {
	// This is a bit of a hack. We need to copy the artifacts from the cache volume to a directory. Without this
	// copy, we're not able to export the artifacts from the container.
	return a.client.Container().
		From("alpine:latest").
		WithMountedCache("/artifacts", a.artifacts).
		WithExec([]string{"cp", "-r", filepath.Clean(filepath.Join("artifacts", runID)), "/out"}).
		Directory("/out")
}
