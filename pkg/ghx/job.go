package ghx

import (
	"fmt"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/idgen"
	"github.com/aweris/gale/internal/log"
)

// JobRunner is the runner that executes the job.
type JobRunner struct {
	RunID      string        // RunID is the run id of the job run.
	Job        core.Job      // Job is the job to be executed.
	context    *gctx.Context // context is the expression context for the job run.
	taskRunner TaskRunner    // taskRunner is the main task taskRunner that executes the job and keeps the execution information.
}

func (r JobRunner) Run(ctx *gctx.Context) (bool, core.Conclusion, error) {
	return r.taskRunner.Run(ctx)
}

// planJob plans the job and returns the job runner.
func planJob(rc *gctx.Context, job core.Job) (*JobRunner, error) {
	runID, err := idgen.GenerateJobRunID()
	if err != nil {
		return nil, err
	}

	runner := &JobRunner{
		RunID:   runID,
		Job:     job,
		context: rc,
	}

	// step task executors that execute the steps
	var (
		setupFns = make([]TaskRunFn, 0)
		pre      = make([]TaskRunner, 0)
		main     = make([]TaskRunner, 0)
		post     = make([]TaskRunner, 0)
	)

	for idx, step := range job.Steps {
		if step.ID == "" {
			step.ID = fmt.Sprintf("%d", idx)
		}

		sr, err := NewStep(step)
		if err != nil {
			return nil, err
		}

		preFn := newTaskPreRunFnForStep(step)
		postFn := newTaskPostRunFnForStep(step)

		// if step implements setup hook, add the setup function to the setupFns slice to be executed
		// by the setup task taskRunner.
		if setup, ok := sr.(SetupHook); ok {
			setupFns = append(setupFns, setup.setup())
		}

		// if step implements pre hook, add the pre task taskRunner to the tasks slice.
		if hook, ok := sr.(PreHook); ok {
			opt := TaskOpts{
				ConditionalFn: hook.preCondition(),
				PreRunFn:      preFn,
				PostRunFn:     postFn,
			}
			pre = append(pre, NewTaskRunner(getStepName("Pre", step), hook.pre(), opt))
		}

		// main task options
		opt := TaskOpts{
			ConditionalFn: sr.condition(),
			PreRunFn:      preFn,
			PostRunFn:     postFn,
		}

		// main tasks starts after pre tasks. so index is step index + len(steps)
		prefix := ""
		if step.Name == "" {
			prefix = "Run"
		}
		main = append(main, NewTaskRunner(getStepName(prefix, step), sr.main(), opt))

		if hook, ok := sr.(PostHook); ok {
			opt := TaskOpts{
				ConditionalFn: hook.postCondition(),
				PreRunFn:      preFn,
				PostRunFn:     postFn,
			}
			post = append(post, NewTaskRunner(getStepName("Post", step), hook.post(), opt))
		}
	}

	var tasks = make([]TaskRunner, 0)

	tasks = append(tasks, NewTaskRunner("Set up job", setup(runner, setup(runner, setupFns...))))
	tasks = append(tasks, pre...)
	tasks = append(tasks, main...)
	tasks = append(tasks, post...)
	tasks = append(tasks, NewTaskRunner("Complete job", complete(runner)))

	runFn := func(ctx *gctx.Context) (core.Conclusion, error) {
		for _, te := range tasks {
			run, conclusion, err := te.Run(ctx)

			// no need to continue if the task taskRunner did not run.
			if !run {
				continue
			}

			if err != nil {
				log.Errorf(te.Name, "error", err.Error())
			}

			// set the job status to the conclusion of the job status is success and the conclusion is not success.
			if runner.context.Job.Status == core.ConclusionSuccess && conclusion != runner.context.Job.Status {
				runner.context.SetJobStatus(conclusion)
			}
		}

		totalSize := 0
		outputs := make(map[string]string, len(runner.Job.Outputs))

		for k, v := range runner.Job.Outputs {
			val, err := expression.NewString(v).Eval(runner.context)
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

			outputs[k] = val
		}

		// According to Github Action docs, The total of all outputs in a workflow run can be a maximum of 50 MB.
		// We'll check the size of the outputs and log a warning if it's bigger than 50MB.
		//
		// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idoutputs
		if totalSize > 50*MB {
			log.Warnf("Total size of the outputs is bigger than 50MB", "size", fmt.Sprintf("%dMB", totalSize/MB))
		}

		return runner.context.Job.Status, nil
	}

	// main task taskRunner that executes the job
	opt := TaskOpts{
		ConditionalFn: nil,
		PreRunFn:      newTaskPreRunFnForJob(runID, job),
		PostRunFn:     newTaskPostRunFnForJob(),
	}
	runner.taskRunner = NewTaskRunner(fmt.Sprintf("Job: %s", job.Name), runFn, opt)

	return runner, nil
}

// setup returns a task taskRunner function that will be executed by the task taskRunner for the setup step.
func setup(_ *JobRunner, setupFns ...TaskRunFn) TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		for _, setupFn := range setupFns {
			conclusion, err := setupFn(ctx)
			if err != nil {
				return conclusion, err
			}
		}

		return core.ConclusionSuccess, nil
	}
}

// MB is the megabyte size in bytes. It'll be used to check size of the job outputs. This just to increase the
// readability of the code.
const MB = 1024 * 1024

// complete returns a task taskRunner function that will be executed by the task taskRunner for the complete step.
func complete(r *JobRunner) TaskRunFn {
	return func(ctx *gctx.Context) (core.Conclusion, error) {
		log.Infof("Complete", "job", r.Job.Name, "conclusion", r.context.Job.Status)

		return core.ConclusionSuccess, nil
	}
}

func newTaskPreRunFnForJob(runID string, job core.Job) TaskPreRunFn {
	return func(ctx *gctx.Context) error {
		ctx.SetJob(
			&core.JobRun{
				RunID:      runID,
				Job:        job,
				Conclusion: "",
				Outcome:    "",
				Outputs:    make(map[string]string),
			},
		)

		return nil
	}
}

func newTaskPostRunFnForJob() TaskPostRunFn {
	return func(ctx *gctx.Context) (err error) {
		ctx.UnsetJob()

		return nil
	}
}
