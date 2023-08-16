package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/log"
	"github.com/aweris/gale/tools/ghx/actions"
)

type Runner struct {
	jr        *core.JobRun         // jr is the job run configuration.
	context   *actions.ExprContext // context is the expression context for the job run.
	executor  TaskExecutor         // executor is the main task executor that executes the job and keeps the execution information.
	stepTasks []TaskExecutor       // stepTasks are the step task executors that execute the steps and keep the execution information.
}

func Plan(jr *core.JobRun) (*Runner, error) {
	runner := &Runner{jr: jr}

	// initialize the expression context
	ec, err := actions.NewExprContext()
	if err != nil {
		return nil, err
	}

	runner.context = ec

	// main task executor that executes the job
	runner.executor = NewTaskExecutor(jr.Job.Name, run(runner))

	// step task executors that execute the steps

	// pre allocate the slices to avoid reallocation.
	setupFns := make([]TaskExecutorFn, 0, len(jr.Job.Steps))
	tasks := make([]TaskExecutor, len(jr.Job.Steps)*3)

	for idx, step := range jr.Job.Steps {
		if step.ID == "" {
			step.ID = fmt.Sprintf("%d", idx)
		}

		sr, err := NewStep(runner, step)
		if err != nil {
			return nil, err
		}

		// setup functions are added to the setupFns slice to be executed by the setup task executor.
		setupFns = append(setupFns, sr.setup())

		// pre task is added same index as the step index
		tasks[idx] = NewConditionalTaskExecutor(getStepName("Pre", step), sr.pre(), sr.preCondition())

		// main tasks starts after pre tasks. so index is step index + len(steps)
		prefix := ""
		if step.Name == "" {
			prefix = "Run"
		}
		tasks[len(jr.Job.Steps)+idx] = NewConditionalTaskExecutor(getStepName(prefix, step), sr.main(), sr.mainCondition())

		// post tasks starts after main tasks. so index is step index + (len(steps) * 2)
		tasks[(len(jr.Job.Steps)*2)+idx] = NewConditionalTaskExecutor(getStepName("Post", step), sr.post(), sr.postCondition())
	}

	runner.stepTasks = append(runner.stepTasks, NewTaskExecutor("Set up job", setup(runner, setup(runner, setupFns...))))
	runner.stepTasks = append(runner.stepTasks, tasks...)
	runner.stepTasks = append(runner.stepTasks, NewTaskExecutor("Complete job", complete(runner)))

	return runner, nil
}

func (r *Runner) Run(ctx context.Context) error {
	// run is always true for the main task executor and concussion not important.
	_, _, err := r.executor.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func run(r *Runner) TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		for _, te := range r.stepTasks {
			run, conclusion, err := te.Run(ctx)

			// no need to continue if the task executor did not run.
			if !run {
				continue
			}

			if err != nil {
				log.Errorf(te.Name, "error", err.Error())
			}

			// set the job status to the conclusion of the job status is success and the conclusion is not success.
			if r.context.Job.Status == core.ConclusionSuccess && conclusion != r.context.Job.Status {
				r.context.SetJobStatus(conclusion)
			}
		}

		return core.ConclusionSuccess, nil
	}
}

// setup returns a task executor function that will be executed by the task executor for the setup step.
func setup(r *Runner, setupFns ...TaskExecutorFn) TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		for _, setupFn := range setupFns {
			_, err := setupFn(ctx)
			if err != nil {
				return core.ConclusionFailure, err
			}
		}

		return core.ConclusionSuccess, nil
	}
}

// complete returns a task executor function that will be executed by the task executor for the complete step.
func complete(r *Runner) TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		log.Infof("Complete", "job", r.jr.Job.Name, "conclusion", r.context.Job.Status)

		return core.ConclusionSuccess, nil
	}
}
