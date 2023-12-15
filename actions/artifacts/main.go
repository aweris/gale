package main

import (
	"context"
	"time"
)

func New() *ActionsArtifactService {
	return &ActionsArtifactService{
		Data: dag.CacheVolume("gale-artifact-service"),
	}
}

type ActionsArtifactService struct {
	Data *CacheVolume
}

// Returns the artifact service itself as a service.
func (m *ActionsArtifactService) Service(ctx context.Context) (*Service, error) {
	var (
		cache = m.Data
		opts  = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

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
	runID string,
) *Directory {
	copySH := `#!/bin/sh
	
echo "Copying artifacts for runID: $1"

if [ -d /artifacts/$1/ ]; then
    echo "Copying artifacts for runID: $1"
    cp -av /artifacts/$1/ /exported_artifacts/
else
    echo "Artifacts for runID: $1 not found"
	
	# list of run ids currently in the cache to help with debugging
	echo "List of run ids in the cache:"
	ls -la /artifacts/
	
    mkdir -p /exported_artifacts
fi
`

	return dag.Container().From("alpine:latest").
		WithMountedCache("/artifacts", m.Data, ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithNewFile("/usr/local/bin/copy-artifacts.sh", ContainerWithNewFileOpts{Contents: copySH, Permissions: 0700}).
		WithExec([]string{"copy-artifacts.sh", runID, time.Now().Format(time.RFC3339Nano)}). // current time just to ensure exec is not cached.
		Directory("/exported_artifacts")
}
