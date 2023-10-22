package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

// WorkflowsRunOpts represents the options for running a workflow.
type WorkflowsRunOpts struct {
	Workflow    string  `doc:"The workflow to run." required:"true"`
	Job         string  `doc:"The job name to run. If empty, all jobs will be run."`
	Event       string  `doc:"Name of the event that triggered the workflow. e.g. push" default:"push"`
	EventFile   *File   `doc:"The file with the complete webhook event payload."`
	RunnerImage string  `doc:"The image to use for the runner." default:"ghcr.io/catthehacker/ubuntu:act-latest"`
	RunnerDebug bool    `doc:"Enable debug mode." default:"false"`
	Token       *Secret `doc:"The GitHub token to use for authentication."`
}

// WorkflowRunDirectoryOpts represents the options for exporting a workflow run.
type WorkflowRunDirectoryOpts struct {
	IncludeRepo      bool `doc:"Include the repository source in the exported directory." default:"false"`
	IncludeSecrets   bool `doc:"Include the secrets in the exported directory." default:"false"`
	IncludeEvent     bool `doc:"Include the event file in the exported directory." default:"false"`
	IncludeArtifacts bool `doc:"Include the artifacts in the exported directory." default:"false"`
}

// WorkflowRunConfig represents the configuration for running a workflow.
type WorkflowRunConfig struct {
	*WorkflowsRepoOpts
	*WorkflowsDirOpts
	*WorkflowsRunOpts
}

type WorkflowRun struct {
	Config *WorkflowRunConfig
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

// Result returns executes the workflow run and returns the result.
func (wr *WorkflowRun) Result(ctx context.Context) (string, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return "", err
	}

	var result WorkflowRunReport

	runs := container.Directory("/home/runner/_temp/ghx/runs")

	// runs directory should only have one entry with the workflow run id
	entries, err := runs.Entries(ctx)
	if err != nil {
		return "", err
	}

	wrID := entries[0]

	resultJSON := filepath.Join("/home/runner/_temp/ghx/runs", wrID, "workflow_run.json")

	err = container.File(resultJSON).unmarshalContentsToJSON(ctx, &result)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Workflow %s completed with conclusion %s in %s", result.Name, result.Conclusion, result.Duration), nil
}

// Directory returns the directory of the workflow run information.
func (wr *WorkflowRun) Directory(ctx context.Context, opts WorkflowRunDirectoryOpts) (*Directory, error) {
	container, err := wr.run(ctx)
	if err != nil {
		return nil, err
	}

	runs := container.Directory("/home/runner/_temp/ghx/runs")

	// runs directory should only have one entry with the workflow run id
	entries, err := runs.Entries(ctx)
	if err != nil {
		return nil, err
	}

	wrID := entries[0]

	dir := dag.Directory().WithDirectory("runs", runs)

	if opts.IncludeRepo {
		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/repo", wrID), container.Directory("."))
	}

	if opts.IncludeSecrets {
		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/secrets", wrID), container.Directory("/home/runner/_temp/ghx/secrets"))
	}

	if opts.IncludeEvent && wr.Config.EventFile != nil {
		dir = dir.WithFile(fmt.Sprintf("runs/%s/event.json", wrID), container.File(filepath.Join("/home", "runner", "work", "_temp", "_github_workflow", "event.json")))
	}

	if opts.IncludeArtifacts {
		container = dag.Container().From("alpine:latest").
			WithMountedCache("/artifacts", dag.Source().ArtifactService().CacheVolume()).
			WithExec([]string{"cp", "-r", fmt.Sprintf("/artifacts/%s", wrID), "/exported_artifacts"})

		dir = dir.WithDirectory(fmt.Sprintf("runs/%s/artifacts", wrID), container.Directory("/exported_artifacts"))
	}

	return dir, nil
}

func (wr *WorkflowRun) run(ctx context.Context) (*Container, error) {
	container, err := wr.container(ctx)
	if err != nil {
		return nil, err
	}

	// loading request scoped configs

	// configure workflow run configuration
	container = container.With(wr.Config.configure)

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

func (wr *WorkflowRun) container(ctx context.Context) (*Container, error) {
	container := dag.Container().From(wr.Config.RunnerImage)

	// set github token as secret if provided
	if wr.Config.Token != nil {
		container = container.WithSecretVariable("GITHUB_TOKEN", wr.Config.Token)
	}

	// configure internal components
	container = container.With(dag.Source().Ghx().Binary)
	container = container.With(dag.Source().ArtifactService().BindAsService)
	container = container.With(dag.Source().ArtifactCacheService().BindAsService)

	// configure repo -- when *Directory can be included in to repo info, we can move source mounting to repo module as well
	var (
		info   = dag.Repo().Info((RepoInfoOpts)(*wr.Config.WorkflowsRepoOpts))
		source = dag.Repo().Source((RepoSourceOpts)(*wr.Config.WorkflowsRepoOpts))
	)

	workdir, err := info.Workdir(ctx)
	if err != nil {
		return nil, err
	}

	container = container.With(info.Configure)
	container = container.WithMountedDirectory(workdir, source).WithWorkdir(workdir)
	container = container.WithEnvVariable("GITHUB_WORKSPACE", workdir)

	// add env variable to the container to indicate container is configured
	container = container.WithEnvVariable("GALE_CONFIGURED", "true")

	return container, nil
}

func (wrc *WorkflowRunConfig) configure(c *Container) *Container {
	container := c

	container = container.WithEnvVariable("GHX_WORKFLOW", wrc.Workflow)
	container = container.WithEnvVariable("GHX_JOB", wrc.Job)
	container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", wrc.WorkflowsDir)

	container = container.WithEnvVariable("GITHUB_EVENT_NAME", wrc.Event)

	if wrc.EventFile != nil {
		container = container.WithMountedFile("/home/runner/_temp/_github_workflow/event.json", wrc.EventFile)
	}

	if wrc.RunnerDebug {
		container = container.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	return container
}
