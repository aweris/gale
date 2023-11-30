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

	return ctr.WithMountedCache(path, cache, opts).WithEnvVariable("GHX_HOME", path)
}
