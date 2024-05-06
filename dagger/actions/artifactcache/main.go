package main

import (
	"context"
)

func New() *ActionsArtifactcacheService {
	return &ActionsArtifactcacheService{
		CacheVol: dag.CacheVolume("gale-artifact-cache-service"),
	}
}

type ActionsArtifactcacheService struct {
	CacheVol *CacheVolume
}

// Returns the artifact service itself as a service.
func (m *ActionsArtifactcacheService) Service(ctx context.Context) (*Service, error) {
	var (
		cache = m.CacheVol
		opts  = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	ctr, err := new(source).Container(ctx)
	if err != nil {
		return nil, err
	}

	// configure cache volume
	ctr = ctr.WithMountedCache("/cache", cache, opts).WithEnvVariable("CACHE_DIR", "/cache")

	// configure service port
	ctr = ctr.WithExposedPort(8081).WithEnvVariable("PORT", "8081")

	// configure service run command
	ctr = ctr.WithExec([]string{"go", "run", "."})

	return ctr.AsService(), nil
}

// Binds the artifact service as a service to the given container and configures ACTIONS_RUNTIME_URL and
// ACTIONS_RUNTIME_TOKEN to allow the artifact service to communicate with the github actions runner.
func (m *ActionsArtifactcacheService) BindAsService(
	// context to use for binding the artifact service.
	ctx context.Context,
	// container to bind the artifact service to.
	ctr *Container,
) (*Container, error) {
	service, err := m.Service(ctx)
	if err != nil {
		return nil, err
	}

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Port: 8081, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	// bind the service to the container
	ctr = ctr.WithServiceBinding("artifact-cache-service", service)

	// set the runtime url and token
	ctr = ctr.WithEnvVariable("ACTIONS_CACHE_URL", endpoint).WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token")

	return ctr, nil
}
