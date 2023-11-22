package main

import (
	"context"
)

type Workflows struct{}

// List returns a list of workflows and their jobs with the given options.
func (w *Workflows) List(
	// context to use for the operation
	ctx context.Context,
	// source directory of the repository
	source *Directory,
	// workflows directory (default: .github/workflows)
	workflowsDir Optional[string],
) (List, error) {
	var workflows []Workflow

	walkFn := func(ctx context.Context, path string, file *File) (bool, error) {
		workflow, err := loadWorkflow(ctx, path, file)
		if err != nil {
			return false, err
		}

		workflows = append(workflows, *workflow)

		return true, nil
	}

	err := walkWorkflowDir(ctx, source, workflowsDir.GetOr(".github/workflows"), walkFn)
	if err != nil {
		return List{}, err
	}

	return List{Workflows: workflows}, nil
}

// Get returns a workflow.
func (w *Workflows) Get(
	// context to use for the operation
	ctx context.Context,
	// source directory of the repository
	source *Directory,
	// workflow name or path
	workflow string,
	// workflows directory (default: .github/workflows)
	workflowsDir Optional[string],
) (*Workflow, error) {
	workflows, err := w.List(ctx, source, workflowsDir)
	if err != nil {
		return nil, err
	}

	return workflows.Get(workflow)
}
