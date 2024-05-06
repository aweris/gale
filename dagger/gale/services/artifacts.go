package services

import (
	"context"
	"time"

	"dagger/gale/internal/dagger"
	"dagger/gale/tools/gobuild"
)

func (s *Services) WithArtifactsService(ctx context.Context, ctr *dagger.Container) (*dagger.Container, error) {
	binary, err := gobuild.Build(
		ctx,
		s.dag,
		gobuild.Workdir("services/artifacts"),
		gobuild.Include("services/artifacts/**/*.go", "services/artifacts/go.*", "libs/common/**/*.go", "libs/common/go.*"),
	)
	if err != nil {
		return nil, err
	}

	var (
		cache = s.dag.CacheVolume("gale-artifact-service")
		opts  = dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}
	)

	svc := s.dag.Container().From("golang:latest").
		WithMountedCache("/artifacts", cache, opts).WithEnvVariable("ARTIFACT_DIR", "/artifacts").
		WithExposedPort(8080).WithEnvVariable("PORT", "8080").
		WithFile("/usr/local/bin/artifacts", binary, dagger.ContainerWithFileOpts{Permissions: 0777}).
		WithExec([]string{"/usr/local/bin/artifacts"}).
		AsService()

	endpoint, err := svc.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: 8080, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	// bind the service to the container
	ctr = ctr.WithServiceBinding("artifact-service", svc)

	// set the runtime url and token
	ctr = ctr.WithEnvVariable("ACTIONS_RUNTIME_URL", endpoint).WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token")

	return ctr, nil
}

func (s *Services) Artifacts(runID string) *dagger.Directory {
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

	return s.dag.Container().From("alpine:latest").
		WithMountedCache("/artifacts", s.dag.CacheVolume("gale-artifact-service"), dagger.ContainerWithMountedCacheOpts{Sharing: dagger.Shared}).
		WithNewFile("/usr/local/bin/copy-artifacts.sh", dagger.ContainerWithNewFileOpts{Contents: copySH, Permissions: 0700}).
		WithExec([]string{"copy-artifacts.sh", runID, time.Now().Format(time.RFC3339Nano)}). // current time just to ensure exec is not cached.
		Directory("/exported_artifacts")
}
