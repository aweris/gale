package ghx

import (
	"context"
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/log"
)

type JobRunner struct {
	jr        *core.JobRun   // jr is the job run configuration.
	context   *ExprContext   // context is the expression context for the job run.
	executor  TaskExecutor   // executor is the main task executor that executes the job and keeps the execution information.
	stepTasks []TaskExecutor // stepTasks are the step task executors that execute the steps and keep the execution information.
}

func Plan(jr *core.JobRun) (*JobRunner, error) {
	runner := &JobRunner{jr: jr}

	// initialize the expression context
	ec, err := NewExprContext()
	if err != nil {
		return nil, err
	}

	runner.context = ec

	// main task executor that executes the job
	runner.executor = NewTaskExecutor(jr.Job.Name, run(runner))

	// step task executors that execute the steps
	var (
		setupFns = make([]TaskExecutorFn, 0)
		pre      = make([]TaskExecutor, 0)
		main     = make([]TaskExecutor, 0)
		post     = make([]TaskExecutor, 0)
	)

	for idx, step := range jr.Job.Steps {
		if step.ID == "" {
			step.ID = fmt.Sprintf("%d", idx)
		}

		sr, err := NewStep(runner, step)
		if err != nil {
			return nil, err
		}

		// if step implements setup hook, add the setup function to the setupFns slice to be executed
		// by the setup task executor.
		if setup, ok := sr.(SetupHook); ok {
			setupFns = append(setupFns, setup.setup())
		}

		// if step implements pre hook, add the pre task executor to the tasks slice.
		if hook, ok := sr.(PreHook); ok {
			// pre task is added same index as the step index
			pre = append(pre, NewConditionalTaskExecutor(getStepName("Pre", step), hook.pre(), hook.preCondition()))
		}

		// main tasks starts after pre tasks. so index is step index + len(steps)
		prefix := ""
		if step.Name == "" {
			prefix = "Run"
		}
		main = append(main, NewConditionalTaskExecutor(getStepName(prefix, step), sr.main(), sr.condition()))

		if hook, ok := sr.(PostHook); ok {
			post = append(post, NewConditionalTaskExecutor(getStepName("Post", step), hook.post(), hook.postCondition()))
		}
	}

	runner.stepTasks = append(runner.stepTasks, NewTaskExecutor("Set up job", setup(runner, setup(runner, setupFns...))))
	runner.stepTasks = append(runner.stepTasks, pre...)
	runner.stepTasks = append(runner.stepTasks, main...)
	runner.stepTasks = append(runner.stepTasks, post...)
	runner.stepTasks = append(runner.stepTasks, NewTaskExecutor("Complete job", complete(runner)))

	return runner, nil
}

func (r *JobRunner) Run(ctx context.Context) error {
	// run is always true for the main task executor and concussion not important.
	_, _, err := r.executor.Run(ctx)
	if err != nil {
		return err
	}

	return nil
}

func run(r *JobRunner) TaskExecutorFn {
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
func setup(_ *JobRunner, setupFns ...TaskExecutorFn) TaskExecutorFn {
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

// MB is the megabyte size in bytes. It'll be used to check size of the job outputs. This just to increase the
// readability of the code.
const MB = 1024 * 1024

// complete returns a task executor function that will be executed by the task executor for the complete step.
func complete(r *JobRunner) TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		totalSize := 0

		for k, v := range r.jr.Job.Outputs {
			val, err := expression.NewString(v).Eval(r.context)
			if err != nil {
				return core.ConclusionFailure, err
			}

			log.Debugf("Evaluated output", "key", k, "value", val)

			// According to Github Action docs, Outputs are Unicode strings, and can be a maximum of 1 MB in size.
			// We'll check the size of the output and log a warning if it's bigger than 1MB.
			//
			// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idoutputs
			if len(val) > 1*MB {
				log.Warnf("Size of the output is bigger than 1MB", "key", k, "size", fmt.Sprintf("%dMB", len(val)/MB))
			}

			totalSize += len(val)
		}

		// According to Github Action docs, The total of all outputs in a workflow run can be a maximum of 50 MB.
		// We'll check the size of the outputs and log a warning if it's bigger than 50MB.
		//
		// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idoutputs
		if totalSize > 50*MB {
			log.Warnf("Total size of the outputs is bigger than 50MB", "size", fmt.Sprintf("%dMB", totalSize/MB))
		}

		log.Infof("Complete", "job", r.jr.Job.Name, "conclusion", r.context.Job.Status)

		return core.ConclusionSuccess, nil
	}
}
