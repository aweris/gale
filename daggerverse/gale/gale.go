package main

import (
	"context"
	"fmt"
)

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct{}

func (g *Gale) runner() *Runner {
	return &Runner{}
}

func (g *Gale) repo() *Repo {
	return &Repo{}
}

// List returns a list of workflows and their jobs with the given options.
func (g *Gale) List(
	// context to use for the operation
	ctx context.Context,
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// Path to the workflows' directory. (default: .github/workflows)
	workflowsDir Optional[string],
) (string, error) {
	// load repository information
	info, err := g.repo().Info(ctx, source, repo, branch, tag)
	if err != nil {
		return "", err
	}

	// load workflows
	workflows := dag.Workflows().List(info.Source, WorkflowsListOpts{WorkflowsDir: workflowsDir.GetOr("")})

	// return string representation of the workflows
	return workflows.String(ctx)
}

// Run runs the workflow with the given options.
func (g *Gale) Run(
	// context to use for the operation
	ctx context.Context,
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// Path to the workflows directory. (default: .github/workflows)
	workflowsDir Optional[string],
	// External workflow file to run.
	workflowFile Optional[*File],
	// Name of the workflow to run.
	workflow Optional[string],
	// Name of the job to run. If empty, all jobs will be run.
	job Optional[string],
	// Name of the event that triggered the workflow. e.g. push
	event Optional[string],
	// File with the complete webhook event payload.
	eventFile Optional[*File],
	// Container to use for the runner(default: ghcr.io/catthehacker/ubuntu:act-latest).
	container Optional[*Container],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token to use for authentication.
	token Optional[*Secret],
) (*WorkflowRun, error) {

	// load repository information
	info, err := g.repo().Info(ctx, source, repo, branch, tag)
	if err != nil {
		return nil, err
	}

	runner, err := g.runner().Container(ctx, container, info)
	if err != nil {
		return nil, err
	}

	wp := ""
	wf, ok := workflowFile.Get()
	if !ok {
		workflow, ok := workflow.Get()
		if !ok {
			return nil, fmt.Errorf("workflow or workflow file must be provided")
		}

		// load workflows
		workflows := dag.Workflows().List(info.Source, WorkflowsListOpts{WorkflowsDir: workflowsDir.GetOr("")})

		w := workflows.Get(workflow)

		wp, err = w.Path(ctx)
		if err != nil {
			return nil, err
		}

		wf = w.Src()
	}

	return &WorkflowRun{
		Runner: runner,
		Config: WorkflowRunConfig{
			WorkflowFile: wf,
			Workflow:     wp,
			Job:          job.GetOr(""),
			Event:        event.GetOr("push"),
			EventFile:    eventFile.GetOr(dag.Directory().WithNewFile("event.json", "{}").File("event.json")),
			RunnerDebug:  runnerDebug.GetOr(false),
			Token:        token.GetOr(nil),
		},
	}, nil
}
