package main

import (
	"fmt"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/core"
	"github.com/aweris/gale/ghx/task"
	"github.com/aweris/gale/internal/log"
)

var (
	_ Step        = new(StepAction)
	_ PreRunHook  = new(StepAction)
	_ PreHook     = new(StepAction)
	_ PostHook    = new(StepAction)
	_ PostRunHook = new(StepAction)
	_ SetupHook   = new(StepAction)
)

// StepAction is a step that runs an action.
type StepAction struct {
	container *dagger.Container
	Step      core.Step
	Action    core.CustomAction
}

func (s *StepAction) setup() task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		path, err := ctx.GetActionsPath()
		if err != nil {
			return core.ConclusionFailure, err
		}

		ca, err := LoadActionFromSource(ctx.Context, ctx.Dagger.Client, s.Step.Uses, path)
		if err != nil {
			return core.ConclusionFailure, err
		}

		// update the step action with the loaded action
		s.Action = *ca

		log.Info(fmt.Sprintf("Download action repository '%s'", s.Step.Uses))

		if s.Action.Meta.Runs.Using == core.ActionRunsUsingDocker {
			var (
				image        = ca.Meta.Runs.Image
				workspace    = ctx.Github.Workspace
				workspaceDir = ctx.Dagger.Client.Host().Directory(workspace)
			)

			switch {
			case image == "Dockerfile":
				s.container = ctx.Dagger.Client.Container().Build(ca.Dir)
			case strings.HasPrefix(image, "docker://"):
				s.container = ctx.Dagger.Client.Container().From(strings.TrimPrefix(image, "docker://"))
			default:
				// This should never happen. Adding it for safety.
				return core.ConclusionFailure, fmt.Errorf("invalid docker image: %s", image)
			}

			// add repository to the container
			s.container = s.container.WithMountedDirectory(workspace, workspaceDir).WithWorkdir(workspace)
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) preRun(stage core.StepStage) task.PreRunFn {
	return func(ctx *context.Context) error {
		ctx.SetAction(&s.Action)

		return ctx.SetStep(
			&core.StepRun{
				Step:    s.Step,
				Stage:   stage,
				Outputs: make(map[string]string),
				State:   make(map[string]string),
			},
		)
	}
}

func (s *StepAction) postRun() task.PostRunFn {
	return func(ctx *context.Context, result task.Result) {
		ctx.UnsetStep(context.RunResult(result))
		ctx.UnsetAction()
	}
}

func (s *StepAction) preCondition() task.ConditionalFn {
	return func(ctx *context.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PreCondition()
		if !run {
			return false, "", nil
		}

		return evalCondition(condition, ctx)
	}
}

func (s *StepAction) pre() task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PreEntrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16, core.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Pre)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}

func (s *StepAction) condition() task.ConditionalFn {
	return func(ctx *context.Context) (bool, core.Conclusion, error) {
		return evalCondition(s.Step.If, ctx)
	}
}

func (s *StepAction) main() task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.Entrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16, core.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Main)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}

func (s *StepAction) postCondition() task.ConditionalFn {
	return func(ctx *context.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PostCondition()
		if !run {
			return false, "", nil
		}

		return evalCondition(condition, ctx)
	}
}

func (s *StepAction) post() task.RunFn {
	return func(ctx *context.Context) (core.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PostEntrypoint)
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16, core.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Post)
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}
