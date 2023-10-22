package main

import (
	"fmt"

	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/core"
	"github.com/aweris/gale/ghx/task"
)

// Step is an internal interface that defines contract for steps.
type Step interface {
	// condition returns the function that checks if the main execution condition is met.
	condition() task.ConditionalFn

	// main returns the function that executes the main execution logic.
	main() task.RunFn
}

// SetupHook is the interface that defines contract for steps capable of performing a setup task.
type SetupHook interface {
	// setup returns the function that sets up the step before execution.
	setup() task.RunFn
}

// PreHook is the interface that defines contract for steps capable of performing a pre execution task.
type PreHook interface {
	// preCondition returns the function that checks if the pre execution condition is met.
	preCondition() task.ConditionalFn

	// pre returns the function that executes the pre execution logic just before the main execution.
	pre() task.RunFn
}

// PostHook is the interface that defines contract for steps capable of performing a post execution task.
type PostHook interface {
	// postCondition returns the function that checks if the post execution condition is met.
	postCondition() task.ConditionalFn

	// post returns the function that executes the post execution logic just after the main execution.
	post() task.RunFn
}

// PreRunHook is the interface that defines contract for steps capable of performing a pre run task. Pre run task is
// executed before the step is executed for each stage.
type PreRunHook interface {
	preRun(stage core.StepStage) task.PreRunFn
}

// PostRunHook is the interface that defines contract for steps capable of performing a post run task. Post run task is
// executed after the step is executed for each stage.
type PostRunHook interface {
	postRun() task.PostRunFn
}

// TODO: separate files for each step type
// TODO: add support for step pre and post run fns
// TODO: if pre and post run fn is missing use default fns

// NewStep creates a new step from the given step configuration.
func NewStep(s core.Step) (Step, error) {
	var step Step

	switch s.Type() {
	case core.StepTypeAction:
		step = &StepAction{Step: s}
	case core.StepTypeRun:
		step = &StepRun{Step: s}
	case core.StepTypeDocker:
		step = &StepDocker{Step: s}
	default:
		return nil, fmt.Errorf("unknown step type: %s", s.Type())
	}

	return step, nil
}

func newTaskPreRunFnForStep(stage core.StepStage, step core.Step) task.PreRunFn {
	return func(ctx *context.Context) error {
		return ctx.SetStep(
			&core.StepRun{
				Step:    step,
				Stage:   stage,
				Outputs: make(map[string]string),
				State:   make(map[string]string),
			},
		)
	}
}

func newTaskPostRunFnForStep() task.PostRunFn {
	return func(ctx *context.Context, result task.Result) {
		ctx.UnsetStep(context.RunResult(result))
	}
}

func executeStep(ctx *context.Context, executor Executor, continueOnError bool) (core.Conclusion, error) {
	// execute the step
	if err := executor.Execute(ctx); err != nil {
		if continueOnError {
			ctx.SetStepResults(core.ConclusionSuccess, core.ConclusionFailure)

			return core.ConclusionSuccess, nil
		}

		ctx.SetStepResults(core.ConclusionFailure, core.ConclusionFailure)

		return core.ConclusionFailure, err
	}

	// update the step outputs
	ctx.SetStepResults(core.ConclusionSuccess, core.ConclusionSuccess)

	return core.ConclusionSuccess, nil
}
