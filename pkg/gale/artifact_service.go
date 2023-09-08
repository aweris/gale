package gale

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
	"github.com/aweris/gale/pkg/data"
)

var _ helpers.WithContainerFuncHook = new(ArtifactService)

// ArtifactService is the dagger service definitions for the services/artifact directory.
type ArtifactService struct {
	client    *dagger.Client
	container *dagger.Container
	data      *data.CacheVolume
	alias     string
	port      string
}

// NewArtifactService creates a new artifact service.
func NewArtifactService(cache *data.CacheVolume) *ArtifactService {
	v := version.GetVersion()

	tag := v.GitVersion

	container := config.Client().Container().From("ghcr.io/aweris/gale/services/artifact:" + tag)

	// port configuration
	container = container.WithEnvVariable("PORT", "8080").WithExposedPort(8080)

	// stateful data configuration

	container = container.With(cache.WithContainerFunc()).WithEnvVariable("ARTIFACTS_DIR", data.ArtifactsPath())

	return &ArtifactService{
		client:    config.Client(),
		container: container,
		data:      cache,
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
