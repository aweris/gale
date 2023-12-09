package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type WorkflowExecutionPlanner struct {
	// information about the repository.
	Repo *RepoInfo

	// workflows in the repository.
	Workflows *Workflows

	// options specific to the workflow run.
	RunOpts *WorkflowRunOpts

	// options specific to the runner.
	RunnerOpts *RunnerOpts

	// options specific to the event.
	EventOpts *EventOpts

	// options specific to the workflow secrets.
	SecretOpts *SecretOpts
}

func NewWorkflowExecutionPlanner(
	repo *RepoInfo,
	workflows *Workflows,
	runOpts *WorkflowRunOpts,
	runnerOpts *RunnerOpts,
	eventOpts *EventOpts,
	secretOpts *SecretOpts,
) *WorkflowExecutionPlanner {
	return &WorkflowExecutionPlanner{
		Repo:       repo,
		Workflows:  workflows,
		RunOpts:    runOpts,
		RunnerOpts: runnerOpts,
		EventOpts:  eventOpts,
		SecretOpts: secretOpts,
	}
}

func (wep *WorkflowExecutionPlanner) Plan(ctx context.Context) (*WorkflowExecutor, error) {
	var workflow *Workflow

	workflow, err := getWorkflow(ctx, wep.Workflows, wep.RunOpts.WorkflowFile, wep.RunOpts.Workflow)
	if err != nil {
		return nil, err
	}

	jobs, err := wep.jobs(workflow)
	if err != nil {
		return nil, err
	}

	return &WorkflowExecutor{
		plan:     wep,
		runID:    uuid.New().String(),
		repo:     wep.Repo,
		workflow: workflow,
		runner:   NewRunner(wep.Repo, workflow, wep.RunnerOpts, wep.EventOpts, wep.SecretOpts),
		jobs:     jobs,
		jrs:      make(map[string]*JobRun),
	}, nil
}

// jobs returns a filtered list of jobs required for execution in a workflow run, sorted by dependency order.
func (wep *WorkflowExecutionPlanner) jobs(workflow *Workflow) ([]*Job, error) {
	var (
		opts    = wep.RunOpts
		jobs    = make(map[string]Job)
		order   = make([]*Job, 0, len(workflow.Jobs))
		visited = make(map[string]bool)

		visitFn func(name string) error
	)

	// initialize map of jobs to work around missing map support in the dagger
	for _, job := range workflow.Jobs {
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
		order = append(order, &job)

		return nil
	}

	if opts.Job != "" {
		if err := visitFn(opts.Job); err != nil {
			return nil, err
		}
	} else {
		for _, job := range workflow.Jobs {
			if err := visitFn(job.JobID); err != nil {
				return nil, err
			}
		}
	}

	if len(order) == 0 {
		return nil, fmt.Errorf("failed to find %s job in the workflow", opts.Job)
	}

	return order, nil
}

// getWorkflow returns the workflow with the given options. IF workflowFile is provided, it will be used. Otherwise,
// workflow will be loaded from the repository source with the given options.
func getWorkflow(ctx context.Context, workflows *Workflows, file *File, workflow string) (*Workflow, error) {
	// FIXME: when dagger supports accepting common input/output types like Custom structs or interfaces from different
	//  modules, we can refactor this to accept a common Workflow type instead of two different options.

	if file != nil {
		return workflows.loadWorkflow(ctx, "", file)
	}

	if workflow == "" {
		return nil, fmt.Errorf("workflow or workflow file must be provided")
	}

	return workflows.Get(ctx, workflow)
}
