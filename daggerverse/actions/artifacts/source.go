package main

import "context"

// source is a Dagger module for managing source code of the project.
type source struct{}

// Code returns the source code of the artifact service.
func (m *source) Code() *Directory {
	return dag.Host().Directory(root(), HostDirectoryOpts{
		Include: []string{
			"common/**/*.go",
			"common/go.*",
			"daggerverse/actions/artifacts/src/*.go",
			"daggerverse/actions/artifacts/src/go.*",
		},
	})
}

// GoMod returns the go.mod file of the artifact service.
func (m *source) GoMod() *File {
	return m.Code().Directory("daggerverse/actions/artifacts/src").File("go.mod")
}

// GoVersion returns the Go version of the artifact service.
func (m *source) GoVersion(ctx context.Context) (string, error) {
	return GoVersion(ctx, m.GoMod())
}

// MountedCode returns the source code of the artifact service mounted in a container at /src and
// sets the working directory to /src/services/artifact.
func (m *source) MountedCode(c *Container) *Container {
	return c.WithMountedDirectory("/src", m.Code()).WithWorkdir("/src/daggerverse/actions/artifacts/src")
}

func (m *source) Container(ctx context.Context) (*Container, error) {
	version, err := m.GoVersion(ctx)
	if err != nil {
		return nil, err
	}

	return goBase(version).With(m.MountedCode).WithExec([]string{"go", "mod", "download"}), nil
}
