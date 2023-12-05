package main

import (
	"context"
	"fmt"
	"path/filepath"
)

type WorkflowRun struct {
	// options for the workflow run.
	Opts *WorkflowRunOpts

	// unique ID of the run.
	RunID string

	// the workflow to run.
	Workflow *Workflow

	// the workflow run report.
	Report *WorkflowRunReport

	// job runs for this workflow run.
	JobRuns []*JobRun
}

type JobRun struct {
	// the job run context for this job run.
	Job Job

	// container for this job run.
	Ctr *Container

	// the directory containing the job run data.
	Data *Directory

	// the report file for this job run.
	Report *JobRunReport

	// the log file for this job run.
	LogFile *File
}

// FIXME: implement sync and directory methods properly. These are just quick hacks to develop the rest of the code.

// Returns all job run logs as a single file.
func (wr *WorkflowRun) Log(ctx context.Context) (*File, error) {
	if len(wr.JobRuns) == 0 {
		return nil, fmt.Errorf("no job runs found")
	}

	var logs string

	for _, jr := range wr.JobRuns {
		contents, err := jr.LogFile.Contents(ctx)
		if err != nil {
			return nil, err
		}

		logs += contents
		logs += "\n"
	}

	logs += fmt.Sprintf("Complete workflow=\"%s\" conclusion=\"%s\" duration=\"%s\"\n", wr.Workflow.Name, wr.Report.Conclusion, wr.Report.Duration)

	return dag.Directory().WithNewFile("logs", logs).File("logs"), nil
}

// Returns the container for the given job id. If there is only one job in the workflow run, then job id is not
// required.
func (wr *WorkflowRun) Sync(
	// job id to return the container for. Only required if there is more than one job in the workflow.
	jobID Optional[string],
) (*Container, error) {
	if len(wr.JobRuns) == 0 {
		return nil, fmt.Errorf("no job runs found")
	}

	if len(wr.JobRuns) == 1 {
		return wr.JobRuns[0].Ctr, nil
	}

	job, ok := jobID.Get()
	if !ok {
		return nil, fmt.Errorf("there are %v job runs in this workflow, please specify a job id", len(wr.JobRuns))
	}

	for _, jr := range wr.JobRuns {
		if jr.Job.JobID == job {
			return jr.Ctr, nil
		}
	}

	return nil, fmt.Errorf("job with id %s not found in workflow run", job)
}

// Returns the directory containing the workflow run data.
func (wr *WorkflowRun) Data() *Directory {
	data := dag.Directory()

	// add workflow file
	data = data.WithFile("run/workflow.yaml", wr.Workflow.Src)

	// add event file
	data = data.WithFile("run/event.json", wr.Opts.EventFile)

	// add workflow run report file
	data = data.WithFile("run/workflow_run.json", wr.Report.File)

	// add job data
	for _, jr := range wr.JobRuns {
		data = data.WithDirectory(filepath.Join("run/jobs", jr.Job.JobID), jr.Data)
	}

	// FIXME: this is currently not working master branch of dagger, only works on v0.9.3 but that version doesn't
	//  have the features I used currently. Validate artifact export later.
	// if includeArtifacts.GetOr(false) {
	//	artifacts := dag.ActionsArtifactService().Artifacts(ActionsArtifactServiceArtifactsOpts{RunID: wr.RunID})
	//
	//	data = data.WithDirectory("artifacts", artifacts)
	// }

	return data
}
