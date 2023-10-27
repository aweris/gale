package main

import (
	"context"
	"fmt"
)

// Source is a Dagger module for managing source code of the project.
type Source struct{}

// Repo returns the source code of the repository.
func (m *Source) Repo() *Directory {
	return dag.Host().Directory(root())
}

func (m *Source) Ghx() *GhxSource {
	return &GhxSource{}
}

func (m *Source) ArtifactService() *ArtifactServiceSource {
	return &ArtifactServiceSource{}
}

func (m *Source) ArtifactCacheService() *ArtifactCacheServiceSource {
	return &ArtifactCacheServiceSource{}
}

// GhxSource represents the source code of the ghx module.
type GhxSource struct{}

// Code returns the source code of the ghx module.
func (m *GhxSource) Code() *Directory {
	return dag.Host().Directory(root(), HostDirectoryOpts{
		Include: []string{
			"common/**/*.go",
			"common/go.*",
			"ghx/**/*.go",
			"ghx/go.*",
		},
	})
}

// GoMod returns the go.mod file of the ghx module.
func (m *GhxSource) GoMod() *File {
	return m.Code().Directory("ghx").File("go.mod")
}

// GoVersion returns the Go version of the ghx module.
func (m *GhxSource) GoVersion(ctx context.Context) (string, error) {
	return GoVersion(ctx, m.GoMod())
}

// MountedCode returns the source code of the ghx module mounted in a container at /src and sets the working directory
// to /src/ghx.
func (m *GhxSource) MountedCode(c *Container) *Container {
	return c.WithMountedDirectory("/src", m.Code()).WithWorkdir("/src/ghx")
}

// Binary adds the ghx binary to the given container and adds binary to the PATH environment variable.
func (m *GhxSource) Binary(ctx context.Context, container *Container) (*Container, error) {
	version, err := m.GoVersion(ctx)
	if err != nil {
		return nil, err
	}

	source, err := goBase(version).
		With(m.MountedCode).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "build", "-o", "bin/ghx", "."}).
		Sync(ctx)
	if err != nil {
		return nil, err
	}

	binary := source.File("bin/ghx")

	if size, err := binary.Size(ctx); err != nil || size == 0 {
		return nil, fmt.Errorf("failed to build ghx: %w", err)
	}

	container = container.WithFile("/usr/local/bin/ghx", binary, ContainerWithFileOpts{Permissions: 0777})

	path, err := container.EnvVariable(ctx, "PATH")
	if err != nil {
		return nil, err
	}

	return container.WithEnvVariable("PATH", fmt.Sprintf("%s:/usr/local/bin", path)), nil
}

// ArtifactServiceSource represents the source code of the artifact service.
type ArtifactServiceSource struct{}

// Code returns the source code of the artifact service.
func (m *ArtifactServiceSource) Code() *Directory {
	return dag.Host().Directory(root(), HostDirectoryOpts{
		Include: []string{
			"common/**/*.go",
			"common/go.*",
			"services/artifact/**/*.go",
			"services/artifact/go.*",
		},
	})
}

// GoMod returns the go.mod file of the artifact service.
func (m *ArtifactServiceSource) GoMod() *File {
	return m.Code().Directory("services/artifact").File("go.mod")
}

// GoVersion returns the Go version of the artifact service.
func (m *ArtifactServiceSource) GoVersion(ctx context.Context) (string, error) {
	return GoVersion(ctx, m.GoMod())
}

// MountedCode returns the source code of the artifact service mounted in a container at /src and
// sets the working directory to /src/services/artifact.
func (m *ArtifactServiceSource) MountedCode(c *Container) *Container {
	return c.WithMountedDirectory("/src", m.Code()).WithWorkdir("/src/services/artifact")
}

func (m *ArtifactServiceSource) CacheVolume() *CacheVolume {
	return dag.CacheVolume("gale-artifact-service")
}

func (m *ArtifactServiceSource) Container(ctx context.Context) (*Container, error) {
	version, err := m.GoVersion(ctx)
	if err != nil {
		return nil, err
	}

	return goBase(version).
		With(m.MountedCode).
		WithExec([]string{"go", "mod", "download"}).
		WithMountedCache("/artifacts", m.CacheVolume(), ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("ARTIFACT_DIR", "/artifacts").
		WithEnvVariable("PORT", "8080").
		WithExposedPort(8080).
		WithExec([]string{"go", "run", "."}), nil
}

func (m *ArtifactServiceSource) BindAsService(ctx context.Context, container *Container) (*Container, error) {
	serviceContainer, err := m.Container(ctx)
	if err != nil {
		return nil, err
	}

	service := serviceContainer.AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Port: 8080, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	return container.
		WithServiceBinding("artifact-service", service).
		WithEnvVariable("ACTIONS_RUNTIME_URL", endpoint).
		WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token"), nil
}

// ArtifactCacheServiceSource represents the source code of the artifact cache service.
type ArtifactCacheServiceSource struct{}

// Code returns the source code of the artifact cache service.
func (m *ArtifactCacheServiceSource) Code() *Directory {
	return dag.Host().Directory(root(), HostDirectoryOpts{
		Include: []string{
			"common/**/*.go",
			"common/go.*",
			"services/artifactcache/**/*.go",
			"services/artifactcache/go.*",
		},
	})
}

// GoMod returns the go.mod file of the artifact cache service.
func (m *ArtifactCacheServiceSource) GoMod() *File {
	return m.Code().Directory("services/artifactcache").File("go.mod")
}

// GoVersion returns the Go version of the artifact cache service.
func (m *ArtifactCacheServiceSource) GoVersion(ctx context.Context) (string, error) {
	return GoVersion(ctx, m.GoMod())
}

// MountedCode returns the source code of the artifact cache service mounted in a container at /src and
// sets the working directory to /src/services/artifactcache.
func (m *ArtifactCacheServiceSource) MountedCode(c *Container) *Container {
	return c.WithMountedDirectory("/src", m.Code()).WithWorkdir("/src/services/artifactcache")
}

func (m *ArtifactCacheServiceSource) CacheVolume() *CacheVolume {
	return dag.CacheVolume("gale-artifact-cache-service")
}

func (m *ArtifactCacheServiceSource) Container(ctx context.Context) (*Container, error) {
	version, err := m.GoVersion(ctx)
	if err != nil {
		return nil, err
	}

	return goBase(version).
		With(m.MountedCode).
		WithExec([]string{"go", "mod", "download"}).
		WithMountedCache("/cache", m.CacheVolume(), ContainerWithMountedCacheOpts{Sharing: Shared}).
		WithEnvVariable("CACHE_DIR", "/cache").
		WithEnvVariable("PORT", "8081").
		WithExposedPort(8081).
		WithExec([]string{"go", "run", "."}), nil
}

func (m *ArtifactCacheServiceSource) BindAsService(ctx context.Context, container *Container) (*Container, error) {
	serviceContainer, err := m.Container(ctx)
	if err != nil {
		return nil, err
	}

	service := serviceContainer.AsService()

	endpoint, err := service.Endpoint(ctx, ServiceEndpointOpts{Port: 8081, Scheme: "http"})
	if err != nil {
		return nil, err
	}

	// normalise the endpoint to remove the trailing slash -- dagger returns the endpoint without the trailing slash
	// but the github actions expects the endpoint with the trailing slash
	endpoint = endpoint + "/"

	return container.
		WithServiceBinding("artifact-cache-service", service).
		WithEnvVariable("ACTIONS_CACHE_URL", endpoint).
		WithEnvVariable("ACTIONS_RUNTIME_TOKEN", "token"), nil
}
