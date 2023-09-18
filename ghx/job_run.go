package main

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/core"
	"github.com/aweris/gale/ghx/expression"
	"github.com/aweris/gale/ghx/idgen"
	"github.com/aweris/gale/ghx/task"
	"github.com/aweris/gale/internal/log"
)

// planJob plans the job and returns the job runner.
func planJob(job core.Job) ([]*task.Runner, error) {
	// step task executors that execute the steps
	var (
		setupFns = make([]task.RunFn, 0)
		pre      = make([]task.Runner, 0)
		main     = make([]task.Runner, 0)
		post     = make([]task.Runner, 0)
	)

	for idx, step := range job.Steps {
		if step.ID == "" {
			step.ID = fmt.Sprintf("%d", idx)
		}

		sr, err := NewStep(step)
		if err != nil {
			return nil, err
		}

		// if step implements setup hook, add the setup function to the setupFns slice to be executed
		// by the setup task taskRunner.
		if setup, ok := sr.(SetupHook); ok {
			setupFns = append(setupFns, setup.setup())
		}

		// if step implements pre hook, add the pre task taskRunner to the tasks slice.
		if hook, ok := sr.(PreHook); ok {
			preRunFn := newTaskPreRunFnForStep(core.StepStagePre, step)
			if pre, ok := sr.(PreRunHook); ok {
				preRunFn = pre.preRun(core.StepStagePre)
			}

			postRunFn := newTaskPostRunFnForStep()
			if post, ok := sr.(PostRunHook); ok {
				postRunFn = post.postRun()
			}
			opt := task.Opts{
				ConditionalFn: hook.preCondition(),
				PreRunFn:      preRunFn,
				PostRunFn:     postRunFn,
			}
			pre = append(pre, task.New(getStepName("Pre", step), hook.pre(), opt))
		}

		preRunFn := newTaskPreRunFnForStep(core.StepStageMain, step)
		if pre, ok := sr.(PreRunHook); ok {
			preRunFn = pre.preRun(core.StepStageMain)
		}

		postRunFn := newTaskPostRunFnForStep()
		if post, ok := sr.(PostRunHook); ok {
			postRunFn = post.postRun()
		}

		// main task options
		opt := task.Opts{
			ConditionalFn: sr.condition(),
			PreRunFn:      preRunFn,
			PostRunFn:     postRunFn,
		}

		// main tasks starts after pre tasks. so index is step index + len(steps)
		prefix := ""
		if step.Name == "" {
			prefix = "Run"
		}
		main = append(main, task.New(getStepName(prefix, step), sr.main(), opt))

		if hook, ok := sr.(PostHook); ok {
			preRunFn := newTaskPreRunFnForStep(core.StepStagePost, step)
			if pre, ok := sr.(PreRunHook); ok {
				preRunFn = pre.preRun(core.StepStagePost)
			}

			postRunFn := newTaskPostRunFnForStep()
			if post, ok := sr.(PostRunHook); ok {
				postRunFn = post.postRun()
			}

			opt := task.Opts{
				ConditionalFn: hook.postCondition(),
				PreRunFn:      preRunFn,
				PostRunFn:     postRunFn,
			}
			post = append(post, task.New(getStepName("Post", step), hook.post(), opt))
		}
	}

	var tasks = make([]task.Runner, 0)

	tasks = append(tasks, task.New("Set up job", setup(setupFns...)))
	tasks = append(tasks, pre...)
	tasks = append(tasks, main...)
	tasks = append(tasks, post...)
	tasks = append(tasks, task.New("Complete job", complete()))

	runFn := func(ctx *context.Context) (core.Conclusion, error) {
		for _, te := range tasks {
			result, err := te.Run(ctx)

			// no need to continue if the task taskRunner did not run.
			if !result.Ran {
				continue
			}

			if err != nil {
				log.Errorf(te.Name, "error", err.Error())
			}

			// set the job status to the conclusion of the job status is success and the conclusion is not success.
			if ctx.Job.Status == core.ConclusionSuccess && result.Conclusion != ctx.Job.Status {
				ctx.Job.Status = result.Conclusion
			}
		}

		totalSize := 0
		outputs := make(map[string]string, len(ctx.Execution.JobRun.Job.Outputs))

		for k, v := range ctx.Execution.JobRun.Job.Outputs {
			val := expression.NewString(v).Eval(ctx)

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

		// TODO: refactor this later into properly. It's added just to make results available for the context.

		ctx.SetJobResults(ctx.Job.Status, ctx.Job.Status, outputs)

		return ctx.Job.Status, nil
	}

	runners := make([]*task.Runner, 0)
	matrices := job.Strategy.Matrix.GenerateCombinations()

	if len(matrices) > 0 {
		for _, matrix := range matrices {
			var values []string

			for k, v := range matrix {
				values = append(values, fmt.Sprintf("%s:%s", k, v))
			}

			sb := strings.Builder{}

			sb.WriteString("Job: ")
			sb.WriteString(job.Name)
			sb.WriteRune('(')
			sb.WriteString(strings.Join(values, ","))
			sb.WriteRune(')')

			runner := task.New(sb.String(), runFn, task.Opts{
				ConditionalFn: newTaskConditionalFnForJob(job),
				PreRunFn:      newTaskPreRunFnForJob(job, matrix),
				PostRunFn:     newTaskPostRunFnForJob(),
			})

			runners = append(runners, &runner)
		}
	} else {
		// task runner options for the job
		opt := task.Opts{
			ConditionalFn: newTaskConditionalFnForJob(job),
			PreRunFn:      newTaskPreRunFnForJob(job),
			PostRunFn:     newTaskPostRunFnForJob(),
		}

		runner := task.New(fmt.Sprintf("Job: %s", job.Name), runFn, opt)

		runners = append(runners, &runner)
	}

	return runners, nil
}

// setup returns a task taskRunner function that will be executed by the task taskRunner for the setup step.
func setup(setupFns ...task.RunFn) task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		// To accurately evaluate conditions like always() or failure(), context inherits the workflow status.
		// This inheritance allow us to evaluate the conditions like always() or failure() correctly. However,
		// we need to reset the status of the job to success before running the steps. Otherwise, the
		// setup steps will not run if the workflow status is failure.
		//
		// To work around this, we're setting the job status to success as a first step of the setup task taskRunner.
		// This will allow us to reset the job status to success after job condition check and before running the
		// actual job steps.
		ctx.Job.Status = core.ConclusionSuccess

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
func complete() task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		log.Infof("Complete", "job", ctx.Execution.JobRun.Job.Name, "conclusion", ctx.Job.Status)

		return core.ConclusionSuccess, nil
	}
}

func newTaskConditionalFnForJob(job core.Job) task.ConditionalFn {
	return func(ctx *context.Context) (bool, core.Conclusion, error) {
		return evalCondition(job.If, ctx)
	}
}

// newTaskPreRunFnForJob returns a task pre run function that will be executed by the task taskRunner for the job. The
// matrix parameter is optional. If it's provided, first matrix combination will be set to the job run.
func newTaskPreRunFnForJob(job core.Job, matrix ...core.MatrixCombination) task.PreRunFn {
	return func(ctx *context.Context) error {
		runID, err := idgen.GenerateJobRunID()
		if err != nil {
			return fmt.Errorf("failed to generate job run id: %w", err)
		}

		jr := &core.JobRun{RunID: runID, Job: job, Outputs: make(map[string]string)}

		if len(matrix) > 0 {
			jr.Matrix = matrix[0]
		}

		return ctx.SetJob(jr)
	}
}

func newTaskPostRunFnForJob() task.PostRunFn {
	return func(ctx *context.Context) (err error) {
		ctx.UnsetJob()

		return nil
	}
}
