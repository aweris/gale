package task

import (
	"fmt"
	"time"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/core"
)

// Status is the phase of the lifecycle that object currently in
type Status string

const (
	StatusQueued     Status = "queued"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
)

// RunFn is the function that will be executed by the task taskRunner.
//
// The return values are:
//   - conclusion: conclusion of the execution
//   - err: error if any
type RunFn func(ctx *context.Context) (conclusion core.Conclusion, err error)

// ConditionalFn is the function that determines if the task should be executed. If the task should not be
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
type ConditionalFn func(ctx *context.Context) (run bool, conclusion core.Conclusion, err error)

// PreRunFn is the function that will be executed before the task is executed. If the function returns an error,
// the task will not be executed. PreRunFn is useful for tasks that need to perform some actions before the execution
// starts.
type PreRunFn func(ctx *context.Context) (err error)

// PostRunFn is the function that will be executed after the task is executed  PostRunFn is useful for tasks that
// need to perform some actions after the execution ends.
type PostRunFn func(ctx *context.Context, result Result)

// Runner is a task runner that runs a task and keeps status, conclusion and timing information about
// the execution.
type Runner struct {
	Name        string        // Name of the execution
	Status      Status        // Status of the execution
	runFn       RunFn         // runFn is the function to be executed
	conditionFn ConditionalFn // conditionFn is the function that determines if the task should be executed
	preFn       PreRunFn      // preFn is the function that will be executed before the task is executed
	postFn      PostRunFn     // postFn is the function that will be executed after the task is executed
}

// Result is the result of the task execution. It is using context.RunResult to avoid circular dependency while keeping
// a result type in the task package as well.
type Result context.RunResult

// Opts is the options that can be used to configure a task.
type Opts struct {
	PreRunFn      PreRunFn
	PostRunFn     PostRunFn
	ConditionalFn ConditionalFn
}

// New creates a new task taskRunner. Optionally, it can be configured with the given options. Only first
// option is used if multiple options are provided.
func New(name string, fn RunFn, opts ...Opts) Runner {
	var opt Opts
	if len(opts) > 0 {
		opt = opts[0]
	}

	return Runner{
		Name:        name,
		Status:      StatusQueued,
		runFn:       fn,
		conditionFn: opt.ConditionalFn,
		preFn:       opt.PreRunFn,
		postFn:      opt.PostRunFn,
	}
}

// Run runs the task and updates the status, conclusion and timing information.
func (t *Runner) Run(ctx *context.Context) (Result, error) {
	var (
		result    = Result{Ran: true}
		startedAt = time.Now()
	)

	t.Status = StatusInProgress

	// run preFn if any
	if t.preFn != nil {
		if err := t.preFn(ctx); err != nil {
			result.Conclusion = core.ConclusionFailure
			result.Duration = time.Since(startedAt)

			return result, err
		}
	}

	// add postFn if any
	if t.postFn != nil {
		// ensure postFn is executed even if there is an error
		defer func() {
			t.postFn(ctx, result)
		}()
	}

	if t.conditionFn != nil {
		run, conclusion, err := t.conditionFn(ctx)
		if err != nil {
			result.Conclusion = core.ConclusionFailure
			result.Duration = time.Since(startedAt)

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

		var err error

		// run the task update named return values
		result.Conclusion, err = t.runFn(ctx)
		if err != nil {
			result.Conclusion = core.ConclusionFailure
			result.Duration = time.Since(startedAt)

			return result, err
		}
	} else if result.Conclusion != "" && result.Conclusion != core.ConclusionSuccess {
		// if the task should not be executed, log the conclusion
		log.Info(fmt.Sprintf("%s (%s)", t.Name, result.Conclusion))
	}

	// update the duration
	result.Duration = time.Since(startedAt)

	// set the task completion
	t.Status = StatusCompleted

	// return the named return values from the task
	return result, nil
}
