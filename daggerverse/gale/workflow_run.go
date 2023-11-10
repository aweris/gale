package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

type WorkflowRun struct {
	// Base container to use for the workflow run.
	Runner *RunnerContainer

	// Configuration of the workflow run.
	Config WorkflowRunConfig
}

// WorkflowRunConfig holds the configuration of a workflow run.
type WorkflowRunConfig struct {
	// Path to the workflow directory.
	WorkflowsDir string

	// WorkflowFile is external workflow file to run.
	WorkflowFile *File

	// Workflow to run.
	Workflow string

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
func (wr *WorkflowRun) Sync() (*Container, error) {
	return wr.run()
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
	container, err := wr.run()
	if err != nil {
		return nil, err
	}

	dir := dag.Directory().WithDirectory("run", container.Directory("/home/runner/_temp/ghx/run"))

	if includeRepo.GetOr(false) {
		dir = dir.WithDirectory("repo", container.Directory("."))
	}

	if includeSecrets.GetOr(false) {
		dir = dir.WithDirectory("secrets", container.Directory("/home/runner/_temp/ghx/secrets"))
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

		container = dag.Container().From("alpine:latest").
			WithMountedCache("/artifacts", dag.Source().ArtifactService().CacheVolume()).
			WithExec([]string{"cp", "-r", fmt.Sprintf("/artifacts/%s", report.RunID), "/exported_artifacts"})

		dir = dir.WithDirectory("artifacts", container.Directory("/exported_artifacts"))
	}

	return dir, nil
}

func (wr *WorkflowRun) run() (*Container, error) {
	container := wr.Runner.Ctr

	// loading request scoped configs

	// configure workflow run configuration
	container = container.With(wr.configure)

	// ghx specific directory configuration
	container = container.WithEnvVariable("GHX_HOME", "/home/runner/_temp/ghx")
	container = container.WithMountedDirectory("/home/runner/_temp/ghx", dag.Directory())
	container = container.WithMountedCache("/home/runner/_temp/ghx/metadata", dag.CacheVolume("gale-metadata"), ContainerWithMountedCacheOpts{Sharing: Shared})
	container = container.WithMountedCache("/home/runner/_temp/ghx/actions", dag.CacheVolume("gale-actions"), ContainerWithMountedCacheOpts{Sharing: Shared})

	// workaround for disabling cache
	container = container.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

	// execute the workflow
	container = container.WithExec([]string{"ghx"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})

	// unloading request scoped configs
	container = container.WithoutEnvVariable("GHX_WORKFLOW")
	container = container.WithoutEnvVariable("GHX_JOB")
	container = container.WithoutEnvVariable("GHX_WORKFLOWS_DIR")

	return container, nil
}

func (wr *WorkflowRun) configure(c *Container) *Container {
	container := c

	// set github token as secret if provided
	if wr.Config.Token != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", wr.Config.Token)
	}

	if wr.Config.WorkflowFile != nil {
		path := "/home/runner/_temp/_github_workflow/.gale/dagger.yaml"

		container = container.WithMountedFile(path, wr.Config.WorkflowFile)
		container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", filepath.Dir(path))

		if wr.Config.Workflow != "" {
			container = container.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow)
		} else {
			container = container.WithEnvVariable("GHX_WORKFLOW", path)
		}
	} else {
		container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", wr.Config.WorkflowsDir)
		container = container.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow)
	}

	container = container.WithEnvVariable("GHX_JOB", wr.Config.Job)

	container = container.WithEnvVariable("GITHUB_EVENT_NAME", wr.Config.Event)

	if wr.Config.EventFile != nil {
		container = container.WithMountedFile("/home/runner/_temp/_github_workflow/event.json", wr.Config.EventFile)
	}

	if wr.Config.RunnerDebug {
		container = container.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	return container
}
