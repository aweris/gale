package repository

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/pkg/gh"
)

// Current returns current repository information from the current working directory.
func Current(ctx context.Context, client *dagger.Client) (*Repo, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	githubRepo, err := gh.CurrentRepository(ctx)
	if err != nil {
		return nil, err
	}

	workflows, err := LoadWorkflows(ctx, client, path)
	if err != nil {
		return nil, err
	}

	return &Repo{
		Repository: githubRepo,
		Path:       path,
		Workflows:  workflows,
	}, nil
}
