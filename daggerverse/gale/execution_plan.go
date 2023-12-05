package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type WorkflowExecutionPlan struct {
	// options for the workflow run.
	opts *WorkflowRunOpts

	// unique ID of the run.
	runID string

	// information about the repository.
	repo *RepoInfo

	// the workflow to run.
	workflow *Workflow

	// map of job runs for this workflow run.
	jrs map[string]*JobRun
}

func NewWorkflowExecutionPlan(ctx context.Context, opts *WorkflowRunOpts) (*WorkflowExecutionPlan, error) {
	var rid = uuid.New().String()

	// load repository information
	info, err := internal.RepoInfo(ctx, opts.Source, opts.Repo, opts.Branch, opts.Tag)
	if err != nil {
		return nil, err
	}

	// set workflow config
	workflow, err := internal.getWorkflow(ctx, info, opts.WorkflowFile, opts.Workflow, opts.WorkflowsDir)
	if err != nil {
		return nil, err
	}

	wep := &WorkflowExecutionPlan{
		opts:     opts,
		runID:    rid,
		repo:     info,
		workflow: workflow,
		jrs:      make(map[string]*JobRun),
	}

	return wep, nil
}

func (wep *WorkflowExecutionPlan) Execute(ctx context.Context) (*WorkflowRun, error) {
	jobs, err := wep.jobs()
	if err != nil {
		return nil, err
	}

	runs := make([]*JobRun, 0, len(jobs))

	rc, err := internal.Runner(wep).Container()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		needs := make([]*JobRun, 0, len(job.Needs))

		for _, need := range job.Needs {
			needs = append(needs, wep.jrs[need])
		}

		jr, err := rc.RunJob(ctx, job, needs...)
		if err != nil {
			return nil, err
		}

		// to keep track of the job runs for able to access them later for dependent jobs
		wep.jrs[job.JobID] = jr

		// add job run to the list of job runs since WorkflowRun is a public type and dagger doesn't support maps yet
		runs = append(runs, jr)
	}

	return &WorkflowRun{Opts: wep.opts, RunID: wep.runID, Workflow: wep.workflow, JobRuns: runs}, nil
}

func (wep *WorkflowExecutionPlan) jobs() ([]Job, error) {
	var (
		jobs    = make(map[string]Job)
		order   = make([]Job, 0, len(wep.workflow.Jobs))
		visited = make(map[string]bool)

		visitFn func(name string) error
	)

	// initialize map of jobs to work around missing map support in the dagger
	for _, job := range wep.workflow.Jobs {
		jobs[job.JobID] = job
	}

	visitFn = func(name string) error {
		if visited[name] {
			return nil
		}

		job, exist := jobs[name]
		if !exist {
			return fmt.Errorf("job %s not found", name)
		}

		visited[name] = true

		for _, dependency := range job.Needs {
			if err := visitFn(dependency); err != nil {
				return err
			}
		}

		// add job to the order slice to keep the order of the jobs
		order = append(order, job)

		return nil
	}

	if wep.opts.Job != "" {
		if err := visitFn(wep.opts.Job); err != nil {
			return nil, err
		}
	} else {
		for _, job := range wep.workflow.Jobs {
			if err := visitFn(job.JobID); err != nil {
				return nil, err
			}
		}
	}

	if len(order) == 0 {
		return nil, fmt.Errorf("failed to find %s job in the workflow", wep.opts.Job)
	}

	return order, nil
}
