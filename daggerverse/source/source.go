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
