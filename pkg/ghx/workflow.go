package ghx

import (
	"errors"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/idgen"
	"github.com/aweris/gale/internal/log"
)

var ErrWorkflowNotFound = errors.New("workflow not found")

// Plan plans the workflow and returns the workflow runner.
func Plan(workflow core.Workflow, job string) (*TaskRunner, error) {
	var (
		order   []string                // order keeps track of job execution order
		visited map[string]bool         // visited keeps track of visited jobs
		visitFn func(name string) error // visitFn is the function that visits the job and its dependencies recursively to create the execution order
	)

	visited = make(map[string]bool)

	visitFn = func(name string) error {
		if _, exist := workflow.Jobs[name]; !exist {
			return fmt.Errorf("job %s not found", name)
		}

		if visited[name] {
			return nil
		}

		visited[name] = true
		for _, need := range workflow.Jobs[name].Needs {
			if err := visitFn(need); err != nil {
				return err
			}
		}

		order = append(order, name)

		return nil
	}

	// if job is specified, visit only that job and its dependencies otherwise visit all jobs
	if job != "" {
		if err := visitFn(job); err != nil {
			return nil, err
		}
	} else {
		for name := range workflow.Jobs {
			if err := visitFn(name); err != nil {
				return nil, err
			}
		}
	}

	// if no jobs found for execution, return error
	if len(order) == 0 {
		return nil, errors.New("no jobs found")
	}

	// log the execution order if there are more than one job
	if len(order) > 1 {
		log.Infof("Job execution order", "jobs", order)
	}

	// runFn is the function that runs the workflow
	runFn := func(ctx *gctx.Context) (core.Conclusion, error) {
		for _, job := range order {
			jm, ok := workflow.Jobs[job]
			if !ok {
				return core.ConclusionFailure, ErrWorkflowNotFound
			}

			runners, err := planJob(jm)
			if err != nil {
				return core.ConclusionFailure, err
			}

			// FIXME: ignoring fail-fast for now. it is always true for now. Fix this later.
			// FIXME: run all runners sequentially for now. Ignoring parallelism. Fix this later.

			for _, runner := range runners {
				_, _, err = runner.Run(ctx)
				if err != nil {
					return core.ConclusionFailure, err
				}
			}
		}

		return core.ConclusionSuccess, nil
	}

	// workflow task options
	opt := TaskOpts{
		PreRunFn:  newTaskPreRunFnForWorkflow(workflow),
		PostRunFn: newTaskPostRunFnForWorkflow(),
	}

	// create the workflow task runner from the runFn and options
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

func newTaskPostRunFnForWorkflow() TaskPostRunFn {
	return func(ctx *gctx.Context) error {
		log.Infof("Complete", "workflow", ctx.Execution.WorkflowRun.Workflow.Name, "conclusion", ctx.Execution.WorkflowRun.Conclusion)
		return nil
	}
}
