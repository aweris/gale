package main

import (
	"context"
	"fmt"
)

const DefaultCacheVolumeKey = "gale-artifact-service"

type ActionsArtifactService struct {
	// ArtifactsVol to use for storing artifacts.
	ArtifactsVol *CacheVolume
}

// Set cache volume for the artifact service.
func (m *ActionsArtifactService) WithCacheVolume(
	// CacheVolume to use for storing artifacts.
	CacheVolume *CacheVolume,
) *ActionsArtifactService {
	m.ArtifactsVol = CacheVolume
	return m
}

// Returns the artifact service itself as a service.
func (m *ActionsArtifactService) Service(ctx context.Context) (*Service, error) {
	var (
		cache = m.ArtifactsVol
		opts  = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	if cache == nil {
		cache = dag.CacheVolume(DefaultCacheVolumeKey)
	}

	ctr, err := new(source).Container(ctx)
	if err != nil {
		return nil, err
	}

	// configure cache volume
	ctr = ctr.WithMountedCache("/artifacts", cache, opts).WithEnvVariable("ARTIFACT_DIR", "/artifacts")

	// configure service port
	ctr = ctr.WithExposedPort(8080).WithEnvVariable("PORT", "8080")

	// configure service run command
	ctr = ctr.WithExec([]string{"go", "run", "."})

	return ctr.AsService(), nil
}

// Binds the artifact service as a service to the given container and configures ACTIONS_RUNTIME_URL and
// ACTIONS_RUNTIME_TOKEN to allow the artifact service to communicate with the github actions runner.
func (m *ActionsArtifactService) BindAsService(
	// context to use for binding the artifact service.
	ctx context.Context,
	// container to bind the artifact service to.
	ctr *Container,
) (*Container, error) {
	service, err := m.Service(ctx)
	if err != nil {
		return nil, err
	}

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Port: 8080, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	// bind the service to the container
	ctr = ctr.WithServiceBinding("artifact-service", service)

	// set the runtime url and token
	ctr = ctr.WithEnvVariable("ACTIONS_RUNTIME_URL", endpoint).WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token")

	return ctr, nil
}

func (m *ActionsArtifactService) Artifacts(
	// run id of the workflow run to get artifacts for. If not provided, all artifacts saved in the cache will be returned.
	runID Optional[string],
) *Directory {
	var (
		cache = m.ArtifactsVol
		opts  = ContainerWithMountedCacheOpts{Sharing: Shared}
		id    = runID.GetOr("")
	)

	if cache == nil {
		cache = dag.CacheVolume(DefaultCacheVolumeKey)
	}

	return dag.Container().From("alpine:latest").
		WithMountedCache("/artifacts", cache, opts).
		WithExec([]string{"cp", "-r", fmt.Sprintf("/artifacts/%s", id), "/exported_artifacts"}).
		Directory("/exported_artifacts")
}
