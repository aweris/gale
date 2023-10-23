package main

import (
	"errors"
	"fmt"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/idgen"
	"github.com/aweris/gale/ghx/task"
)

// planWorkflow plans the workflow and returns the workflow runner.
func planWorkflow(workflow model.Workflow, job string) (*task.Runner, error) {
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
	runFn := func(ctx *context.Context) (model.Conclusion, error) {
		conclusion := model.ConclusionSuccess

		for _, job := range order {
			jm, ok := workflow.Jobs[job]
			if !ok {
				return model.ConclusionFailure, fmt.Errorf("job %s not found", job)
			}

			runners, err := planJob(jm)
			if err != nil {
				return model.ConclusionFailure, err
			}

			// FIXME: ignoring fail-fast for now. it is always true for now. Fix this later.
			// FIXME: run all runners sequentially for now. Ignoring parallelism. Fix this later.

			for _, runner := range runners {
				result, err := runner.Run(ctx)
				if err != nil {
					return model.ConclusionFailure, err
				}

				if conclusion == model.ConclusionSuccess && result.Conclusion != conclusion {
					conclusion = result.Conclusion
				}
			}
		}

		return conclusion, nil
	}

	// workflow task options
	opt := task.Opts{
		PreRunFn:  newTaskPreRunFnForWorkflow(workflow),
		PostRunFn: newTaskPostRunFnForWorkflow(),
	}

	// create the workflow task runner from the runFn and options
	runner := task.New(fmt.Sprintf("Workflow: %s", workflow.Name), runFn, opt)

	return &runner, nil
}

func newTaskPreRunFnForWorkflow(wf model.Workflow) task.PreRunFn {
	return func(ctx *context.Context) error {
		runID, err := idgen.GenerateWorkflowRunID(ctx)
		if err != nil {
			return fmt.Errorf("failed to generate workflow run id: %w", err)
		}

		return ctx.SetWorkflow(
			&model.WorkflowRun{
				RunID:         runID,
				RunNumber:     "1",
				RunAttempt:    "1",
				RetentionDays: "0",
				Workflow:      wf,
				Jobs:          make(map[string]model.JobRun),
			},
		)
	}
}

func newTaskPostRunFnForWorkflow() task.PostRunFn {
	return func(ctx *context.Context, result task.Result) {
		log.Infof("Complete", "workflow", ctx.Execution.WorkflowRun.Workflow.Name, "conclusion", result.Conclusion)

		ctx.UnsetWorkflow(context.RunResult(result))
	}
}
