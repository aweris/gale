package main

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
)

type WorkflowRun struct {
	// Configuration for the workflow run.
	Config WorkflowRunConfig
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
		info = dag.Repo().Info(RepoInfoOpts{
			Source: wr.Config.Source,
			Repo:   wr.Config.Repo,
			Branch: wr.Config.Branch,
			Tag:    wr.Config.Tag,
		})
		source = dag.Repo().Source(RepoSourceOpts{
			Source: wr.Config.Source,
			Repo:   wr.Config.Repo,
			Branch: wr.Config.Branch,
			Tag:    wr.Config.Tag,
		})
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

func (wr *WorkflowRun) configure(c *Container) *Container {
	container := c

	container = container.WithEnvVariable("GHX_WORKFLOW", wr.Config.Workflow)
	container = container.WithEnvVariable("GHX_JOB", wr.Config.Job)
	container = container.WithEnvVariable("GHX_WORKFLOWS_DIR", wr.Config.WorkflowsDir)

	container = container.WithEnvVariable("GITHUB_EVENT_NAME", wr.Config.Event)

	if wr.Config.EventFile != nil {
		container = container.WithMountedFile("/home/runner/_temp/_github_workflow/event.json", wr.Config.EventFile)
	}

	if wr.Config.RunnerDebug {
		container = container.WithEnvVariable("RUNNER_DEBUG", "1")
	}

	return container
}
