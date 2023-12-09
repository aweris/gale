package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type WorkflowRun struct {
	// unique ID of the run.
	RunID string

	// event options for the workflow run.
	Event *EventOpts

	// the workflow to run.
	Workflow *Workflow

	// the workflow run report.
	Report *WorkflowRunReport

	// job runs for this workflow run.
	JobRuns []*JobRun
}

type JobRun struct {
	// the job run context for this job run.
	Job *Job

	// container for this job run.
	Ctr *Container

	// the directory containing the job run data.
	Data *Directory

	// the report file for this job run.
	Report *JobRunReport

	// the log file for this job run.
	LogFile *File
}

// Returns all job run logs as a single file.
func (wr *WorkflowRun) Log(ctx context.Context) (*File, error) {
	if len(wr.JobRuns) == 0 {
		return nil, fmt.Errorf("no job runs found")
	}

	var logs strings.Builder

	for _, jr := range wr.JobRuns {
		contents, err := jr.LogFile.Contents(ctx)
		if err != nil {
			return nil, err
		}

		logs.WriteString(contents)
		logs.WriteString("\n")
	}

	// add summary of the workflow run
	logs.WriteString(fmt.Sprintf("Complete workflow=\"%s\" ", wr.Workflow.Name))
	logs.WriteString(fmt.Sprintf("conclusion=\"%s\" ", wr.Report.Conclusion))
	logs.WriteString(fmt.Sprintf("duration=\"%s\"", wr.Report.Duration))
	logs.WriteString("\n")

	return dag.Directory().WithNewFile("logs", logs.String()).File("logs"), nil
}

// Returns the container for the given job id. If there is only one job in the workflow run, then job id is not required.
func (wr *WorkflowRun) Sync(
	// job id to return the container for. Only required if there is more than one job in the workflow.
	// +optional=true
	jobID string,
) (*Container, error) {
	jobCount := len(wr.JobRuns)

	// decide what to do based on the number of jobs in the workflow run

	// if there in no job in the workflow run, return an error
	if jobCount == 0 {
		return nil, fmt.Errorf("no job runs found")
	}

	// if there is only one job in the workflow run, return the container for that job. No need to specify job id.
	if jobCount == 1 {
		return wr.JobRuns[0].Ctr, nil
	}

	// if there is more than one job in the workflow run, job id is required to pick the right container
	if jobID == "" {
		return nil, fmt.Errorf("there are %v job runs in this workflow, please specify a job id", len(wr.JobRuns))
	}

	// since map type is not supported yet, we have to iterate over the job runs to find the right one
	for _, jr := range wr.JobRuns {
		if jr.Job.JobID == jobID {
			return jr.Ctr, nil
		}
	}

	// if we get here, it means the job id is not found in the workflow run
	return nil, fmt.Errorf("job with id %s not found in workflow run", jobID)
}

// Returns the directory containing the workflow run data.
func (wr *WorkflowRun) Data() *Directory {
	data := dag.Directory()

	// add workflow file
	data = data.WithFile("run/workflow.yaml", wr.Workflow.Src)

	// add event file
	data = data.WithFile("run/event.json", wr.Event.File)

	// add workflow run report file
	data = data.WithFile("run/workflow_run.json", wr.Report.File)

	// add job data
	for _, jr := range wr.JobRuns {
		data = data.WithDirectory(filepath.Join("run/jobs", jr.Job.JobID), jr.Data)
	}

	// add artifacts if any
	data = data.WithDirectory("artifacts", dag.ActionsArtifactService().Artifacts(wr.RunID))

	return data
}
