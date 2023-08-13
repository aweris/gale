package services

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/version"
)

var _ helpers.WithContainerFuncHook = new(ArtifactCacheService)

// ArtifactCacheService is the dagger service definitions for the services/artifact directory.
type ArtifactCacheService struct {
	client        *dagger.Client
	container     *dagger.Container
	artifactcache *dagger.CacheVolume
	alias         string
	port          string
}

// NewArtifactCacheService creates a new artifact service.
func NewArtifactCacheService() *ArtifactCacheService {
	v := version.GetVersion()

	tag := v.GitVersion

	container := config.Client().Container().From("ghcr.io/aweris/gale/services/artifactcache:" + tag)

	// port configuration
	container = container.WithEnvVariable("PORT", "8080").WithExposedPort(8080)

	// stateful data configuration

	cache := config.Client().CacheVolume("gale-artifactcache-service")
	container = container.WithMountedCache("/cache", cache).WithEnvVariable("CACHE_DIR", "/cache")

	return &ArtifactCacheService{
		client:        config.Client(),
		container:     container,
		artifactcache: cache,
		alias:         "artifactcache",
		port:          "8080",
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
