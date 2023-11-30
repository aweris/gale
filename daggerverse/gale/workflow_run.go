package main

import (
	"context"
	"path/filepath"
	"time"
)

type WorkflowRun struct {
	// Context of the workflow run.
	Context *RunContext
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

	rd := container.WithExec([]string{"cp", "-r", wr.Context.getSharedDataMountPath(), "/exported_run"}).Directory("/exported_run")

	dir := dag.Directory().WithDirectory("run", rd.Directory("run"))

	if includeSecrets.GetOr(false) {
		dir = dir.WithDirectory("secrets", rd.Directory("secrets"))
	}

	if includeRepo.GetOr(false) {
		dir = dir.WithDirectory("repo", container.Directory("."))
	}

	if includeEvent.GetOr(false) && wr.Context.Opts.EventFile != nil {
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
	var opts = wr.Context.Opts

	// load repository information
	info, err := internal.repo().Info(ctx, opts.Source, opts.Repo, opts.Branch, opts.Tag)
	if err != nil {
		return nil, err
	}

	rc, err := internal.runner().Container(ctx, info, opts.Container)
	if err != nil {
		return nil, err
	}

	// get runner container and apply run context
	ctr := rc.Ctr.With(wr.Context.ContainerFunc)

	// set workflow config
	w, err := internal.getWorkflow(ctx, info.Source, opts.WorkflowFile, opts.Workflow, opts.WorkflowsDir)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(wr.Context.getSharedDataMountPath(), "run", "workflow.yaml")

	ctr = ctr.WithMountedFile(path, w.Src)
	ctr = ctr.WithEnvVariable("GHX_WORKFLOW", w.Name)
	ctr = ctr.WithEnvVariable("GHX_JOB", opts.Job)

	// workaround for disabling cache
	ctr = ctr.WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano))

	// execute the workflow
	ctr = ctr.WithExec([]string{"ghx"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})

	// unloading request scoped configs
	ctr = ctr.WithoutEnvVariable("GHX_JOB")

	return ctr, nil
}
