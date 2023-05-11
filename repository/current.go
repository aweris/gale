package repository

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
	"github.com/aweris/gale/github/cli"
)

// Current returns current repository information from the current working directory.
func Current(ctx context.Context, client *dagger.Client) (*Repo, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	githubRepo, err := cli.CurrentRepository(ctx)
	if err != nil {
		return nil, err
	}

	workflows, err := loadWorkflows(ctx, client, path)
	if err != nil {
		return nil, err
	}

	dh := filepath.Join(config.DataHome(), strings.TrimPrefix(githubRepo.URL, "https://"))

	return &Repo{
		Repository: githubRepo,
		Path:       path,
		DataHome:   dh,
		Workflows:  workflows,
	}, nil
}
