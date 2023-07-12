package gale

import (
	"context"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/dagger/services"
	"github.com/aweris/gale/internal/model"
)

type ExecResult struct {
	Container *dagger.Container

	// contexts
	github *model.GithubContext

	// services
	artifactService *services.ArtifactService
}

// ExportRunnerDirectory exports the runner directory contains all configuration, logs and artifacts to the host
// machine. This is useful for debugging purposes.
func (r *ExecResult) ExportRunnerDirectory(ctx context.Context, path string) error {
	_, err := r.Container.Directory(containerRunnerPath).Export(ctx, path)
	if err != nil {
		return err
	}

	_, err = r.artifactService.Artifacts(r.github.RunID).Export(ctx, filepath.Join(path, "artifacts"))
	if err != nil {
		return err
	}

	return nil
}

func (g *Gale) Exec(ctx context.Context) (*ExecResult, error) {
	container, err := g.Container()
	if err != nil {
		return nil, err
	}

	result := &ExecResult{
		Container:       container,
		github:          g.github,
		artifactService: g.artifactService,
	}

	result.Container = result.Container.WithExec([]string{"ghx", "run"})

	// we're not interested in the output of the container. We just want to make sure that the container is running
	// we'll get the exit code later from the container
	_, _ = result.Container.ExitCode(ctx)

	exitCode, err := result.Container.Directory(containerRunnerPath).File(runnerExitCodeFile).Contents(ctx)
	if err != nil {
		fmt.Printf("failed to get exit code: %v\n", err)
		return result, err
	}

	if exitCode != "0" {
		return result, fmt.Errorf("exit code: %s", exitCode)
	}

	return result, nil
}
