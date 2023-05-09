package builder

import (
	"dagger.io/dagger"
)

// CreateFn is a function that creates a Dagger container.
type CreateFn func(client *dagger.Client) *dagger.Container

// From initializes this container from a pulled base image. This would be conflicting with Dockerfile and Import
// options.
func (b *Builder) From(from string) *Builder {
	b.createFn = func(client *dagger.Client) *dagger.Container {
		return client.Container().From(from)
	}
	return b
}

// Dockerfile Initializes this container from a Dockerfile build. If no opts are provided, the default dockerfile
// path is './Dockerfile'. This would be conflicting with From and Import options.
func (b *Builder) Dockerfile(dir *dagger.Directory, opts ...dagger.ContainerBuildOpts) *Builder {
	b.createFn = func(client *dagger.Client) *dagger.Container {
		return client.Container().Build(dir, opts...)
	}
	return b
}

// Import reads the container from an OCI tarball. This would be conflicting with From and Dockerfile options.
func (b *Builder) Import(source *dagger.File, opts ...dagger.ContainerImportOpts) *Builder {
	b.createFn = func(client *dagger.Client) *dagger.Container {
		return client.Container().Import(source, opts...)
	}
	return b
}
