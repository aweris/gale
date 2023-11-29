package main

import (
	"context"
	"fmt"
)

type Runner struct{}

// RunnerContainer represents a container to run a Github Actions workflow in.
type RunnerContainer struct {
	// Initialized container to run the workflow in.
	Ctr *Container
}

// Container initializes a new runner container with the given options.
func (r *Runner) Container(
	// context to use for the operation
	ctx context.Context,
	// repository information to use for the runner
	repo *RepoInfo,
	// Container to use for the runner.
	ctr *Container,
) (*RunnerContainer, error) {
	// check if container is already initialized
	if ctr != nil {
		if isContainerInitialized(ctx, ctr) {
			fmt.Println("skipping container initialization, container already initialized")
			fmt.Println("WARNING: given source and repo options are ignored, using the initialized container")

			return &RunnerContainer{Ctr: ctr}, nil
		}
	}

	if ctr == nil {
		ctr = dag.Container().From("ghcr.io/catthehacker/ubuntu:act-latest")
	}

	// configure internal components
	ctr = ctr.With(dag.Ghx().Source().Binary)
	ctr = ctr.With(dag.ActionsArtifactService().BindAsService)
	ctr = ctr.With(dag.ActionsArtifactcacheService().BindAsService)

	// extra services
	ctr = ctr.With(dag.Docker().WithCacheVolume("gale-docker-cache").BindAsService)

	// configure repository
	workdir := fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name)

	ctr = ctr.WithMountedDirectory(workdir, repo.Source).WithWorkdir(workdir)
	ctr = ctr.WithEnvVariable("GH_REPO", repo.NameWithOwner)
	ctr = ctr.WithEnvVariable("GITHUB_WORKSPACE", workdir)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY", repo.NameWithOwner)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY_OWNER", repo.Owner)
	ctr = ctr.WithEnvVariable("GITHUB_REPOSITORY_URL", repo.URL)
	ctr = ctr.WithEnvVariable("GITHUB_REF", repo.Ref)
	ctr = ctr.WithEnvVariable("GITHUB_REF_NAME", repo.RefName)
	ctr = ctr.WithEnvVariable("GITHUB_REF_TYPE", repo.RefType)
	ctr = ctr.WithEnvVariable("GITHUB_SHA", repo.SHA)

	// ghx specific directory configuration -- TODO: refactor this later to be more generic for runners
	var (
		metadata  = "/home/runner/_temp/gale/metadata"
		actions   = "/home/runner/_temp/gale/actions"
		cacheOpts = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	ctr = ctr.WithEnvVariable("GHX_METADATA_DIR", metadata)
	ctr = ctr.WithMountedCache(metadata, dag.CacheVolume("gale-metadata"), cacheOpts)

	ctr = ctr.WithEnvVariable("GHX_ACTIONS_DIR", actions)
	ctr = ctr.WithMountedCache(actions, dag.CacheVolume("gale-actions"), cacheOpts)

	// add env variable to the container to indicate container is configured
	ctr = ctr.WithEnvVariable("GALE_CONFIGURED", "true")

	return &RunnerContainer{Ctr: ctr}, nil
}

// isContainerInitialized checks if the given container is initialized and returns the runner id if it is.
func isContainerInitialized(ctx context.Context, container *Container) bool {
	val, err := container.EnvVariable(ctx, "GALE_CONFIGURED")
	if err != nil {
		return false
	}

	return val == "true"
}
