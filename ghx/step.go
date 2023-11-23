package main

import (
	"fmt"

	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/common/task"
	"github.com/aweris/gale/ghx/context"
)

// Step is an internal interface that defines contract for steps.
type Step interface {
	// condition returns the function that checks if the main execution condition is met.
	condition() task.ConditionalFn[context.Context]

	// main returns the function that executes the main execution logic.
	main() task.RunFn[context.Context]
}

// SetupHook is the interface that defines contract for steps capable of performing a setup task.
type SetupHook interface {
	// setup returns the function that sets up the step before execution.
	setup() task.RunFn[context.Context]
}

// PreHook is the interface that defines contract for steps capable of performing a pre execution task.
type PreHook interface {
	// preCondition returns the function that checks if the pre execution condition is met.
	preCondition() task.ConditionalFn[context.Context]

	// pre returns the function that executes the pre execution logic just before the main execution.
	pre() task.RunFn[context.Context]
}

// PostHook is the interface that defines contract for steps capable of performing a post execution task.
type PostHook interface {
	// postCondition returns the function that checks if the post execution condition is met.
	postCondition() task.ConditionalFn[context.Context]

	// post returns the function that executes the post execution logic just after the main execution.
	post() task.RunFn[context.Context]
}

// PreRunHook is the interface that defines contract for steps capable of performing a pre run task. Pre run task is
// executed before the step is executed for each stage.
type PreRunHook interface {
	preRun(stage model.StepStage) task.PreRunFn[context.Context]
}

// PostRunHook is the interface that defines contract for steps capable of performing a post run task. Post run task is
// executed after the step is executed for each stage.
type PostRunHook interface {
	postRun() task.PostRunFn[context.Context]
}

// TODO: separate files for each step type
// TODO: add support for step pre and post run fns
// TODO: if pre and post run fn is missing use default fns

// NewStep creates a new step from the given step configuration.
func NewStep(s model.Step) (Step, error) {
	var step Step

	switch s.Type() {
	case model.StepTypeAction:
		step = &StepAction{Step: s}
	case model.StepTypeRun:
		step = &StepRun{Step: s}
	case model.StepTypeDocker:
		step = &StepDocker{Step: s}
	default:
		return nil, fmt.Errorf("unknown step type: %s", s.Type())
	}

	return step, nil
}

func newTaskPreRunFnForStep(stage model.StepStage, step model.Step) task.PreRunFn[context.Context] {
	return func(ctx *context.Context) error {
		return ctx.SetStep(
			&model.StepRun{
				Step:    step,
				Stage:   stage,
				Outputs: make(map[string]string),
				State:   make(map[string]string),
			},
		)
	}
}

func newTaskPostRunFnForStep() task.PostRunFn[context.Context] {
	return func(ctx *context.Context, result task.Result) {
		ctx.UnsetStep(context.RunResult(result))
	}
}

func executeStep(ctx *context.Context, executor Executor, continueOnError bool) (model.Conclusion, error) {
	// execute the step
	if err := executor.Execute(ctx); err != nil {
		if continueOnError {
			ctx.SetStepResults(model.ConclusionSuccess, model.ConclusionFailure)

			return model.ConclusionSuccess, nil
		}

		ctx.SetStepResults(model.ConclusionFailure, model.ConclusionFailure)

		return model.ConclusionFailure, err
	}

	// update the step outputs
	ctx.SetStepResults(model.ConclusionSuccess, model.ConclusionSuccess)

	return model.ConclusionSuccess, nil
}
