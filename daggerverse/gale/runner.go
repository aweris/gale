package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Runner struct{}

// RunnerContainer represents a container to run a Github Actions workflow in.
type RunnerContainer struct {
	// Unique identifier for the runner container.
	RunnerID string

	// Initialized container to run the workflow in.
	Ctr *Container
}

// getRunnerContainer returns a runner container for the given container if it is initialized.
func (r *Runner) getRunnerContainer(ctx context.Context, container *Container) *RunnerContainer {
	id := isContainerInitialized(ctx, container)
	if id == "" {
		return nil
	}

	return &RunnerContainer{RunnerID: id, Ctr: container}
}

// Container initializes a new runner container with the given options.
func (r *Runner) Container(
	// context to use for the operation
	ctx context.Context,
	// Image to use for the runner. If --image and --container provided together, --container takes precedence.
	image Optional[string],
	// Container to use for the runner. If --image and --container provided together, --container takes precedence.
	container Optional[*Container],
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
) *RunnerContainer {
	// check if container is already initialized
	if ctr, ok := container.Get(); ok {
		id := isContainerInitialized(ctx, ctr)
		if id != "" {
			println(fmt.Sprintf("skip container initialization, container already initialized with id: %s", id))
			println(fmt.Sprintf("WARNING: given source and repo options are ignored, using the initialized container"))

			return &RunnerContainer{RunnerID: id, Ctr: ctr}
		}
	}

	var (
		id       = uuid.New().String()
		imageVal = image.GetOr("ghcr.io/catthehacker/ubuntu:act-latest")
		ctr      = container.GetOr(dag.Container().From(imageVal))
		info     = getRepoInfo(source, repo, branch, tag)
		path     = getRunnerCacheVolumeMountPath(id)
		cache    = getRunnerCacheVolume(id)
	)

	// configure internal components
	ctr = ctr.With(dag.Source().Ghx().Binary)
	ctr = ctr.With(dag.Source().ArtifactService().BindAsService)
	ctr = ctr.With(dag.Source().ArtifactCacheService().BindAsService)

	// configure repo
	ctr = ctr.With(info.Configure)

	// configure runner cache
	ctr = ctr.WithEnvVariable("GALE_RUNNER_CACHE", path)
	ctr = ctr.WithMountedCache(path, cache, ContainerWithMountedCacheOpts{Sharing: Shared})

	// add env variable to the container to indicate container is configured
	ctr = ctr.WithEnvVariable("GALE_RUNNER_ID", id)
	ctr = ctr.WithEnvVariable("GALE_CONFIGURED", "true")

	return &RunnerContainer{RunnerID: id, Ctr: ctr}
}

func (rc *RunnerContainer) Container() *Container {
	return rc.Ctr
}

func (rc *RunnerContainer) getRunnerCachePath() string {
	return getRunnerCacheVolumeMountPath(rc.RunnerID)
}

// isContainerInitialized checks if the given container is initialized and returns the runner id if it is.
func isContainerInitialized(ctx context.Context, container *Container) string {
	val, err := container.EnvVariable(ctx, "GALE_CONFIGURED")
	if err != nil {
		return ""
	}

	if val != "true" {
		return ""
	}

	runnerID, err := container.EnvVariable(ctx, "GALE_RUNNER_ID")
	if err != nil {
		return ""
	}

	return runnerID
}

// getRunnerCacheVolumeMountPath returns the mount path for the cache volume with the given id.
func getRunnerCacheVolumeMountPath(id string) string {
	return fmt.Sprintf("/home/runner/_temp/gale/%s", id)
}

// getRunnerCacheVolume returns a cache volume for the given id.
func getRunnerCacheVolume(id string) *CacheVolume {
	return dag.CacheVolume(fmt.Sprintf("gale-runner-%s", id))
}
