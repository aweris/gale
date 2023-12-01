package main

import "path/filepath"

// Holds the information about the current run.
type RunContext struct {
	// Unique ID of the run.
	RunID string

	// Options to configure the workflow run.
	Opts *WorkflowRunOpts

	// Cache volume to share data between jobs in the same run and keep the data after the run.
	SharedData *CacheVolume
}

// getSharedDataMountPath returns the path to mount the shared data volume.
func (rc *RunContext) getSharedDataMountPath() string {
	return filepath.Join("/home/runner/_temp/_gale/runs", rc.RunID)
}

func (rc *RunContext) ContainerFunc(ctr *Container) *Container {
	var (
		path  = rc.getSharedDataMountPath()
		cache = rc.SharedData
		opts  = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	ctr = ctr.WithMountedCache(path, cache, opts).WithEnvVariable("GHX_HOME", path)

	// set github token as secret if provided
	if rc.Opts.Token != nil {
		ctr = ctr.WithSecretVariable("GITHUB_TOKEN", rc.Opts.Token)
	}

	// set runner debug mode if enabled
	if rc.Opts.RunnerDebug {
		ctr = ctr.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	// event config
	eventPath := filepath.Join(path, "run", "event.json")

	ctr = ctr.WithEnvVariable("GITHUB_EVENT_NAME", rc.Opts.Event)
	ctr = ctr.WithEnvVariable("GITHUB_EVENT_PATH", eventPath)
	ctr = ctr.WithMountedFile(eventPath, rc.Opts.EventFile)

	// workflow run config
	ctr = ctr.WithEnvVariable("GITHUB_RUN_ID", rc.RunID)
	ctr = ctr.WithEnvVariable("GITHUB_RUN_NUMBER", "1")
	ctr = ctr.WithEnvVariable("GITHUB_RUN_ATTEMPT", "1")
	ctr = ctr.WithEnvVariable("GITHUB_RETENTION_DAYS", "90")

	return ctr
}
