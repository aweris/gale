package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type WorkflowRun struct {
	// Workflow run cache path to mount to runner containers.
	RunCachePath string

	// Workflow run cache volume to share data between jobs in the same workflow run and keep the data after the workflow
	RunCacheVolume *CacheVolume

	// Configuration of the workflow run.
	Config WorkflowRunConfig
}

// WorkflowRunConfig holds the configuration of a workflow run.
type WorkflowRunConfig struct {
	// Base container to use for running the workflow.
	BaseContainer *Container

	// Repository information to use for the workflow run.
	Repo *RepoInfo

	// Workflow to run.
	Workflow *Workflow

	// Job name to run. If empty, all jobs will be run.
	Job string

	// Name of the event that triggered the workflow. e.g. push
	Event string

	// File with the complete webhook event payload.
	EventFile *File

	// Enables debug mode.
	RunnerDebug bool

	// GitHub token to use for authentication.
	Token *Secret
}

// FIXME: add jobs to WorkflowRunReport when dagger supports map type

// WorkflowRunReport represents the result of a workflow run.
type WorkflowRunReport struct {
	Ran           bool   `json:"ran"`            // Ran indicates if the execution ran
	Duration      string `json:"duration"`       // Duration of the execution
	Name          string `json:"name"`           // Name is the name of the workflow
	Path          string `json:"path"`           // Path is the path of the workflow
	RunID         string `json:"run_id"`         // RunID is the ID of the run
	RunNumber     string `json:"run_number"`     // RunNumber is the number of the run
	RunAttempt    string `json:"run_attempt"`    // RunAttempt is the attempt number of the run
	RetentionDays string `json:"retention_days"` // RetentionDays is the number of days to keep the run logs
	Conclusion    string `json:"conclusion"`     // Conclusion is the result of a completed workflow run after continue-on-error is applied
}

// Sync runs the workflow and returns the container that ran the workflow.
func (wr *WorkflowRun) Sync(ctx context.Context) (*Container, error) {
	return wr.run(ctx)
}

// Directory returns the directory of the workflow run information.
func (wr *WorkflowRun) Directory(
	ctx context.Context,
	// Adds the repository source to the exported directory. (default: false)
	includeRepo Optional[bool],
	// Adds the mounted secrets to the exported directory. (default: false)
	includeSecrets Optional[bool],
	// Adds the event file to the exported directory. (default: false)
	includeEvent Optional[bool],
	// Adds the uploaded artifacts to the exported directory. (default: false)
	includeArtifacts Optional[bool],
) (*Directory, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return nil, err
	}

	rd := container.WithExec([]string{"cp", "-r", wr.RunCachePath, "/exported_run"}).Directory("/exported_run")

	dir := dag.Directory().WithDirectory("run", rd.Directory("run"))

	if includeSecrets.GetOr(false) {
		dir = dir.WithDirectory("secrets", rd.Directory("secrets"))
	}

	if includeRepo.GetOr(false) {
		dir = dir.WithDirectory("repo", container.Directory("."))
	}

	if includeEvent.GetOr(false) && wr.Config.EventFile != nil {
		dir = dir.WithFile("event.json", container.File("/home/runner/_temp/_github_workflow/event.json"))
	}

	if includeArtifacts.GetOr(false) {
		var report WorkflowRunReport

		err := unmarshalContentsToJSON(ctx, dir.File("run/workflow_run.json"), &report)
		if err != nil {
			return nil, err
		}

		artifacts := dag.ActionsArtifactService().Artifacts(ActionsArtifactServiceArtifactsOpts{RunID: report.RunID})

		dir = dir.WithDirectory("artifacts", artifacts)
	}

	return dir, nil
}

func (wr *WorkflowRun) run(ctx context.Context) (*Container, error) {
	rc, err := internal.runner().Container(ctx, wr.Config.Repo, wr.Config.BaseContainer)
	if err != nil {
		return nil, err
	}

	var (
		ctr       = rc.Ctr
		id        = uuid.New().String()
		wrPath    = filepath.Join("/home/runner/_temp/_gale/runs", id)
		cache     = dag.CacheVolume(fmt.Sprintf("ghx-run-%s", id))
		cacheOpts = ContainerWithMountedCacheOpts{Sharing: Shared}
	)

	// mount workflow run cache volume
	wr.RunCachePath = wrPath
	wr.RunCacheVolume = cache

	ctr = ctr.WithMountedCache(wrPath, cache, cacheOpts)
	ctr = ctr.WithEnvVariable("GHX_HOME", wrPath)

	// set github token as secret if provided
	if wr.Config.Token != nil {
		ctr = ctr.WithSecretVariable("GITHUB_TOKEN", wr.Config.Token)
	}

	// set runner debug mode if enabled
	if wr.Config.RunnerDebug {
		ctr = ctr.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	// set workflow config
	path := filepath.Join(wrPath, "run", "workflow.yaml")

	ctr = ctr.WithMountedFile(path, wr.Config.Workflow.Src)
	ctr = ctr.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow.Name)
	ctr = ctr.WithEnvVariable("GHX_JOB", wr.Config.Job)

	// event config
	eventPath := filepath.Join(wrPath, "run", "event.json")

	ctr = ctr.WithEnvVariable("GITHUB_EVENT_NAME", wr.Config.Event)
	ctr = ctr.WithEnvVariable("GITHUB_EVENT_PATH", eventPath)
	ctr = ctr.WithMountedFile(eventPath, wr.Config.EventFile)

	// workaround for disabling cache
	ctr = ctr.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

	// execute the workflow
	ctr = ctr.WithExec([]string{"ghx"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})

	// unloading request scoped configs
	ctr = ctr.WithoutEnvVariable("GHX_JOB")
	ctr = ctr.WithoutEnvVariable("GHX_WORKFLOWS_DIR")

	return ctr, nil
}
