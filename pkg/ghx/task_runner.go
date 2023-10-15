package ghx

import (
	"fmt"
	"time"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/log"
)

// TaskRunner is a task taskRunner that runs a task and keeps status, conclusion and timing information about
// the execution.
type TaskRunner struct {
	Name        string            // Name of the execution
	Status      core.Status       // Status of the execution
	runFn       TaskRunFn         // runFn is the function to be executed
	conditionFn TaskConditionalFn // conditionFn is the function that determines if the task should be executed
	preFn       TaskPreRunFn      // preFn is the function that will be executed before the task is executed
	postFn      TaskPostRunFn     // postFn is the function that will be executed after the task is executed
}

type TaskRunResult struct {
	Ran         bool            `json:"ran"`         // Ran indicates if the execution ran
	Conclusion  core.Conclusion `json:"conclusion"`  // Conclusion of the execution
	StartedAt   time.Time       `json:"startedAt"`   // StartedAt time of the execution
	CompletedAt time.Time       `json:"completedAt"` // CompletedAt time of the execution
}

// TaskRunFn is the function that will be executed by the task taskRunner.
//
// The return values are:
//   - conclusion: conclusion of the execution
//   - err: error if any
type TaskRunFn func(ctx *gctx.Context) (conclusion core.Conclusion, err error)

// TaskConditionalFn is the function that determines if the task should be executed. If the task should not be
// executed, the conclusion is returned as well.
//
// The return values are:
//   - run: indicates if the task should be executed
//   - conclusion: conclusion of the execution if the task should not be executed (run = false)
//   - err: error if any
//
// The rules are:
//   - If err is not nil, run is false and conclusion is ignored.
//   - If run is true, conclusion and err are ignored.
//   - If run is false, conclusion is not empty and err is nil, do not run the task and set the conclusion.
//   - If run is false, conclusion is empty and err is nil, invalid task. Ignore it completely. The reason this
//     scenario exist is that the task is not invalid by itself. It is just not applicable to the current context, and
//     we can't determine if it is invalid or not in planning phase.
type TaskConditionalFn func(ctx *gctx.Context) (run bool, conclusion core.Conclusion, err error)

// TaskPreRunFn is the function that will be executed before the task is executed. If the function returns an error,
// the task will not be executed. PreRunFn is useful for tasks that need to perform some actions before the execution
// starts.
type TaskPreRunFn func(ctx *gctx.Context) (err error)

// TaskPostRunFn is the function that will be executed after the task is executed. If the function returns an error,
// the task will be marked as failed. PostRunFn is useful for tasks that need to perform some actions after the
// execution ends.
type TaskPostRunFn func(ctx *gctx.Context) (err error)

// TaskOpts is the options that can be used to configure a task.
type TaskOpts struct {
	PreRunFn      TaskPreRunFn
	PostRunFn     TaskPostRunFn
	ConditionalFn TaskConditionalFn
}

// NewTaskRunner creates a new task taskRunner. Optionally, it can be configured with the given options. Only first
// option is used if multiple options are provided.
func NewTaskRunner(name string, fn TaskRunFn, opts ...TaskOpts) TaskRunner {
	var opt TaskOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	return TaskRunner{
		Name:        name,
		Status:      core.StatusQueued,
		runFn:       fn,
		conditionFn: opt.ConditionalFn,
		preFn:       opt.PreRunFn,
		postFn:      opt.PostRunFn,
	}
}

// Run runs the task and updates the status, conclusion and timing information.
func (t *TaskRunner) Run(ctx *gctx.Context) (result TaskRunResult, err error) {
	result = TaskRunResult{Ran: true, StartedAt: time.Now()}

	t.Status = core.StatusInProgress

	// run preFn if any
	if t.preFn != nil {
		if err := t.preFn(ctx); err != nil {
			result.Conclusion = core.ConclusionFailure
			result.CompletedAt = time.Now()

			return result, err
		}
	}

	if t.conditionFn != nil {
		run, conclusion, err := t.conditionFn(ctx)
		if err != nil {
			result.Conclusion = core.ConclusionFailure
			result.CompletedAt = time.Now()

			return result, err
		}

		result.Ran = run
		result.Conclusion = conclusion
	}

	if result.Ran {
		// create ger group for step
		log.Info(t.Name)
		log.StartGroup()
		defer log.EndGroup()

		// run the task update named return values
		result.Conclusion, err = t.runFn(ctx)
		if err != nil {
			result.Conclusion = core.ConclusionFailure
			result.CompletedAt = time.Now()

			return result, err
		}
	} else if result.Conclusion != "" && result.Conclusion != core.ConclusionSuccess {
		// if the task should not be executed, log the conclusion
		log.Info(fmt.Sprintf("%s (%s)", t.Name, result.Conclusion))
	}

	// run postFn if any
	if t.postFn != nil {
		if err := t.postFn(ctx); err != nil {
			result.Conclusion = core.ConclusionFailure
			result.CompletedAt = time.Now()

			return result, err
		}
	}

	// set the task completion
	result.CompletedAt = time.Now()

	// return the named return values from the task
	return result, err
}
