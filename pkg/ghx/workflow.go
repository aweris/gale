package ghx

import (
	"errors"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/idgen"
)

var ErrWorkflowNotFound = errors.New("workflow not found")

// Plan plans the workflow and returns the workflow runner.
func Plan(workflow core.Workflow, job string) (*TaskRunner, error) {
	runFn := func(ctx *gctx.Context) (core.Conclusion, error) {
		jm, ok := workflow.Jobs[job]
		if !ok {
			return core.ConclusionFailure, ErrWorkflowNotFound
		}

		jr, err := planJob(jm)
		if err != nil {
			return core.ConclusionFailure, err
		}

		_, _, err = jr.Run(ctx)
		if err != nil {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}

	// workflow task options
	opt := TaskOpts{PreRunFn: newTaskPreRunFnForWorkflow(workflow)}

	runner := NewTaskRunner(fmt.Sprintf("Workflow: %s", workflow.Name), runFn, opt)

	return &runner, nil
}

func newTaskPreRunFnForWorkflow(wf core.Workflow) TaskPreRunFn {
	return func(ctx *gctx.Context) error {
		runID, err := idgen.GenerateWorkflowRunID()
		if err != nil {
			return fmt.Errorf("failed to generate workflow run id: %w", err)
		}

		return ctx.SetWorkflow(
			&core.WorkflowRun{
				RunID:         runID,
				RunNumber:     "1",
				RunAttempt:    "1",
				RetentionDays: "0",
				Workflow:      wf,
				Jobs:          make(map[string]core.JobRun),
			},
		)
	}
}
