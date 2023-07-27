package runner

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/tools/ghx/internal/log"
)

type Runner struct {
	jr        *core.JobRun   // jr is the job run configuration.
	executor  TaskExecutor   // executor is the main task executor that executes the job and keeps the execution information.
	stepTasks []TaskExecutor // stepTasks are the step task executors that execute the steps and keep the execution information.
}

func Plan(jr *core.JobRun) (*Runner, error) {
	runner := &Runner{jr: jr}

	// main task executor that executes the job
	runner.executor = NewTaskExecutor(jr.Job.Name, run(runner))

	runner.stepTasks = append(runner.stepTasks, NewTaskExecutor("Set up job", setup(runner)))
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

			// only fail if the step failed and the conclusion is not success.
			if err != nil && conclusion != core.ConclusionSuccess {
				// TODO: handle error, remaining steps should have always() or failure() conditions to run
				//  otherwise they should be cancelled.
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

		log.Info(fmt.Sprintf("Complete job name: %s", r.jr.Job.Name))

		return core.ConclusionSuccess, nil
	}
}

// complete returns a task executor function that will be executed by the task executor for the complete step.
func complete(r *Runner) TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		log.Info(fmt.Sprintf("Complete job name: %s", r.jr.Job.Name))

		return core.ConclusionSuccess, nil
	}
}
