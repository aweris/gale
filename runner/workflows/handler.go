package workflows

import (
	"context"

	"github.com/aweris/gale/runner/jobs"
	"github.com/aweris/gale/runner/state"
)

type Handler struct {
	state *state.WorkflowRunState
}

func NewHandler(state *state.WorkflowRunState) *Handler {
	return &Handler{state: state}
}

func (h *Handler) RunWorkflow(ctx context.Context, workflow string) error {
	// initialize workflow and workflow run
	err := h.state.NewWorkflowRun(workflow)
	if err != nil {
		return err
	}

	for name, job := range h.state.Workflow.Jobs {
		if job.Name == "" {
			job.Name = name
		}

		jobs.NewHandler(h.state.GetJobRunState(job)).RunJob(ctx)
	}

	return nil
}
