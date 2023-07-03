package services

import (
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/version"
)

// ArtifactService is the dagger service definitions for the services/artifact directory.
type ArtifactService struct {
	client    *dagger.Client
	container *dagger.Container
	artifacts *dagger.CacheVolume
	alias     string
	port      string
}

// NewArtifactService creates a new artifact service.
func NewArtifactService(client *dagger.Client) *ArtifactService {
	v := version.GetVersion()

	tag := v.GitVersion

	// If the tag is a dev tag, we'll use the main branch.
	if tag == "v0.0.0-dev" {
		tag = "main"
	}

	container := client.Container().From("ghcr.io/aweris/gale/services/artifact:" + tag)

	// port configuration
	container = container.WithEnvVariable("PORT", "8080").WithExposedPort(8080)

	// stateful data configuration

	cache := client.CacheVolume("gale-artifact-service")
	container = container.WithMountedCache("/artifacts", cache).WithEnvVariable("ARTIFACTS_DIR", "/artifacts")

	return &ArtifactService{
		client:    client,
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

// ServiceBinding returns a container with the artifact service binding and all necessary configurations. The method
// signature is compatible with the dagger.WithContainerFunc type. It can be used to as
// container.With(service.ServiceBinding) to bind the service to the container.
func (a *ArtifactService) ServiceBinding(container *dagger.Container) *dagger.Container {
	container = container.WithServiceBinding(a.alias, a.container)
	container = container.WithEnvVariable("ACTIONS_RUNTIME_URL", fmt.Sprintf("http://%s:%s/", a.alias, a.port))
	container = container.WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token") // dummy token, not used by service

	return container
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
