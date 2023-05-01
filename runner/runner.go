package runner

import (
	"context"
	"fmt"
	"github.com/aweris/gale/gha"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
)

// Runner represents a GitHub Action runner powered by Dagger.
type Runner struct {
	Client    *dagger.Client
	Container *dagger.Container

	ActionsBySource     map[string]*gha.Action
	ActionPathsBySource map[string]string

	events []Event
}

// NewRunner creates a new Runner.
func NewRunner(ctx context.Context, client *dagger.Client) (*Runner, error) {
	// check if there is a pre-built runner image
	path, _ := config.SearchDataFile(filepath.Join(config.DefaultRunnerLabel, config.DefaultRunnerImageTar))
	if path != "" {
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		fmt.Printf("Found pre-built image for %s, importing...\n", config.DefaultRunnerLabel)

		container := client.Container().Import(client.Host().Directory(dir).File(base))

		return &Runner{
			Client:              client,
			Container:           container,
			ActionsBySource:     make(map[string]*gha.Action),
			ActionPathsBySource: make(map[string]string),
		}, nil
	}

	fmt.Printf("No pre-built image found for %s, building a new one...\n", config.DefaultRunnerLabel)

	// Build the runner with the defaults and return it, if there is no pre-built image
	return NewBuilder(client).Build(ctx)
}
