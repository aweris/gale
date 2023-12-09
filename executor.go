package main

import (
	"context"
	"fmt"
	"time"

	"github.com/aweris/gale/common/model"
)

type WorkflowExecutor struct {
	// execution plan for this workflow run.
	plan *WorkflowExecutionPlanner

	// unique ID of the run.
	runID string

	// information about the repository.
	repo *RepoInfo

	// the workflow to run.
	workflow *Workflow

	// the workflow run report.
	runner *Runner

	// jobs needed to run sorted by execution order.
	jobs []*Job

	// map of job runs for this workflow run.
	jrs map[string]*JobRun
}

func (we *WorkflowExecutor) Execute(ctx context.Context) (*WorkflowRun, error) {
	var (
		conclusion = model.ConclusionSuccess
		startedAt  = time.Now()
		runs       = make([]*JobRun, 0, len(we.jobs))
	)

	// get the container instance to run the jobs
	rc, err := we.runner.Container(we.runID)
	if err != nil {
		return nil, err
	}

	for _, job := range we.jobs {
		// find dependent job runs to pass to the job run
		needs := make([]*JobRun, 0, len(job.Needs))

		for _, need := range job.Needs {
			needs = append(needs, we.jrs[need])
		}

		// execute the job run with the given dependencies and current workflow conclusion
		jr, err := rc.RunJob(ctx, job, string(conclusion), needs...)
		if err != nil {
			return nil, err
		}

		// only update the workflow run conclusion if the job run conclusion is not success and the workflow run
		// conclusion is success to avoid overriding the non-success conclusion
		//
		// initial conclusion   job conclusion   final conclusion
		// -------------------------------------------------------
		// success              success          success
		// success              failure          failure
		// failure              success          failure
		if conclusion == model.ConclusionSuccess && jr.Report.Conclusion != model.ConclusionSuccess {
			conclusion = jr.Report.Conclusion
		}

		// to keep track of the job runs for able to access them later for dependent jobs
		we.jrs[job.JobID] = jr

		// add job run to the list of job runs since WorkflowRun is a public type and dagger doesn't support maps yet
		runs = append(runs, jr)
	}

	// create the workflow run report
	report, err := NewWorkflowRunReport(true, we.runID, we.workflow, conclusion, time.Since(startedAt), runs)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow run report: %w", err)
	}

	return &WorkflowRun{
		RunID:    we.runID,
		Workflow: we.workflow,
		Event:    we.plan.EventOpts,
		Report:   report,
		JobRuns:  runs,
	}, nil
}
