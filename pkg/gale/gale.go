package gale

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
)

// RunOpts are the options for the Run function.
type RunOpts struct {
	Repo         string
	Branch       string
	Tag          string
	Commit       string
	WorkflowsDir string
}

// Run runs a job from a workflow.
func Run(ctx context.Context, workflow, job string, opts ...RunOpts) error {
	opt := RunOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	repo, err := core.GetRepository(opt.Repo, core.GetRepositoryOpts{Branch: opt.Branch, Tag: opt.Tag, Commit: opt.Commit})
	if err != nil {
		return err
	}

	workflows, err := repo.LoadWorkflows(ctx, core.RepositoryLoadWorkflowOpts{WorkflowsDir: opt.WorkflowsDir})
	if err != nil {
		return err
	}

	wf, ok := workflows[workflow]
	if !ok {
		return ErrWorkflowNotFound
	}

	jm, ok := wf.Jobs[job]
	if !ok {
		return ErrJobNotFound
	}

	fmt.Printf("Running job %s from workflow %s\n", jm.Name, wf.Name)

	return nil
}
