package gale

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
	"github.com/aweris/gale/pkg/data"
)

var _ helpers.WithContainerFuncHook = new(ArtifactCacheService)

// ArtifactCacheService is the dagger service definitions for the services/artifact directory.
type ArtifactCacheService struct {
	client    *dagger.Client
	container *dagger.Container
	data      *data.CacheVolume
	alias     string
	port      string
}

// NewArtifactCacheService creates a new artifact service.
func NewArtifactCacheService(cache *data.CacheVolume) *ArtifactCacheService {
	var (
		tag   = version.GetVersion().GitVersion
		alias = "artifactcache"
		port  = 8080
	)

	container := config.Client().Container().From("ghcr.io/aweris/gale/services/artifactcache:" + tag)

	// external hostname configuration

	container = container.WithEnvVariable("EXTERNAL_HOSTNAME", alias)

	// port configuration

	container = container.WithEnvVariable("PORT", fmt.Sprintf("%d", port)).WithExposedPort(port)

	// stateful data configuration

	container = container.With(cache.WithContainerFunc()).WithEnvVariable("CACHE_DIR", data.ArtifactsCachePath())

	// debug configuration -- enable debug logging for internal/log package
	if config.Debug() {
		container = container.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	return &ArtifactCacheService{
		client:    config.Client(),
		container: container,
		data:      cache,
		alias:     alias,
		port:      fmt.Sprintf("%d", port),
	}
}

// Container returns the container of the artifact service.
func (a *ArtifactCacheService) Container() *dagger.Container {
	return a.container
}

func (a *ArtifactCacheService) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithServiceBinding(a.alias, a.container).
			WithEnvVariable("ACTIONS_CACHE_URL", fmt.Sprintf("http://%s:%s/", a.alias, a.port)).
			WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token") // dummy token, not used by service
	}
}
