package builder

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
	"github.com/aweris/gale/repository"
)

// Builder represents a builder for creating a GitHub Action runner.
type Builder struct {
	// client is the Dagger client used to create the container.
	client *dagger.Client

	// repository is the path to the repository to be used for the runner export.
	repository string

	// label of the container.
	label string

	// createFn is the function that creates the container.
	createFn CreateFn

	// modifyFns are the functions that modify the container.
	modifyFns []ModifyFn
}

// WithRunnerLabel sets the label for the runner container.
func (b *Builder) WithRunnerLabel(label string) *Builder {
	b.label = label
	return b
}

// NewBuilder creates a new Builder.
func NewBuilder(client *dagger.Client, repo *repository.Repo) *Builder {
	// create a new builder instance with default values
	builder := &Builder{
		client:     client,
		repository: strings.TrimPrefix(repo.URL, "https://"), // use github.com/owner/repo as the repository
		label:      config.DefaultRunnerLabel,
	}

	// Set default container creation and modification functions
	builder.bootstrap()

	return builder
}

func (b *Builder) Build(ctx context.Context) (*dagger.Container, error) {
	// create a container
	container := b.createFn(b.client)

	// apply all modifyFns to the container
	for _, step := range b.modifyFns {
		container = step(container)
	}

	// Set the user to the runner user. User is created in default modifyFns added in NewBuilder.
	container = container.WithUser("runner")

	directory := filepath.Join(config.DataHome(), b.repository, "images", b.label)
	file := filepath.Join(directory, config.DefaultRunnerImageTar)

	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	// Export the container to a tarball in the data home directory($XDG_DATA_HOME/gale/owner/repo/<runner-label>/image.tar).
	// This tarball will be used avoid rebuilding the runner image every time and reduce relying on cache.
	_, err := container.Export(ctx, file)
	if err != nil {
		return nil, err
	}

	return container, err
}
