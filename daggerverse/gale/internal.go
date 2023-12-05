package main

import (
	"context"
	"fmt"
)

// The 'internal' struct acts as the central hub for this module's internal operations. It organizes and streamlines
// internal functionalities, ensuring cohesive and efficient interaction among various components. This design aids
// in maintaining a clean code structure and simplifies the management of internal dependencies.

var internal Internal

type Internal struct{}

func (_ *Internal) WorkflowExecutionPlan(
	ctx context.Context,
	source *Directory,
	repo string,
	tag string,
	branch string,
	workflowsDir string,
	workflowFile Optional[*File],
	workflow Optional[string],
	job Optional[string],
	container Optional[*Container],
	event Optional[string],
	eventFile Optional[*File],
	runnerDebug Optional[bool],
	token Optional[*Secret],
) (*WorkflowExecutionPlan, error) {
	return NewWorkflowExecutionPlan(
		ctx,
		toWorkflowRunOpts(
			source,
			repo,
			tag,
			branch,
			workflowsDir,
			workflowFile,
			workflow,
			job,
			container,
			event,
			eventFile,
			runnerDebug,
			token,
		),
	)
}

func (_ *Internal) RepoInfo(ctx context.Context, source *Directory, repo, tag, branch string) (*RepoInfo, error) {
	return NewRepoInfo(ctx, source, repo, tag, branch)
}

func (_ *Internal) Runner(plan *WorkflowExecutionPlan) *Runner {
	return NewRunner(plan)
}

// getWorkflow returns the workflow with the given options. IF workflowFile is provided, it will be used. Otherwise,
// workflow will be loaded from the repository source with the given options.
func (_ *Internal) getWorkflow(ctx context.Context, info *RepoInfo, file *File, workflow string, dir string) (*Workflow, error) {
	// FIXME: when dagger supports accepting common input/output types like Custom structs or interfaces from different
	//  modules, we can refactor this to accept a common Workflow type instead of two different options.

	if file != nil {
		return info.workflows().loadWorkflow(ctx, "", file)
	}

	if workflow == "" {
		return nil, fmt.Errorf("workflow or workflow file must be provided")
	}

	return info.workflows().Get(ctx, workflow, dir)
}
