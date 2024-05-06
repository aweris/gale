package services

import (
	"context"

	"dagger/gale/internal/dagger"
	"dagger/gale/tools/gobuild"
)

func (s *Services) WithArtifactcacheService(ctx context.Context, ctr *dagger.Container) (*dagger.Container, error) {
	binary, err := gobuild.Build(
		ctx,
		s.dag,
		gobuild.Workdir("services/artifactcache"),
		gobuild.Include("services/artifactcache/**/*.go", "services/artifactcache/go.*", "libs/common/**/*.go", "libs/common/go.*"),
	)
	if err != nil {
		return nil, err
	}

	var (
		cache = s.dag.CacheVolume("gale-artifact-cache-service")
		opts  = dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}
	)

	svc := s.dag.Container().From("golang:latest").
		WithMountedCache("/cache", cache, opts).WithEnvVariable("CACHE_DIR", "/cache").
		WithExposedPort(8081).WithEnvVariable("PORT", "8081").
		WithFile("/usr/local/bin/artifactcache", binary, dagger.ContainerWithFileOpts{Permissions: 0777}).
		WithExec([]string{"/usr/local/bin/artifactcache"}).
		AsService()

	endpoint, err := svc.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 8081, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	// bind the service to the container
	ctr = ctr.WithServiceBinding("artifact-cache-service", svc)

	// set the runtime url and token
	ctr = ctr.WithEnvVariable("ACTIONS_CACHE_URL", endpoint).WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token")

	return ctr, nil
}
