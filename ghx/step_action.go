package main

import (
	"fmt"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/task"
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
	Step      model.Step
	Action    model.CustomAction
}

func (s *StepAction) setup() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		path, err := ctx.GetActionsPath()
		if err != nil {
			return model.ConclusionFailure, err
		}

		ca, err := LoadActionFromSource(ctx.Context, ctx.Dagger.Client, s.Step.Uses, path)
		if err != nil {
			return model.ConclusionFailure, err
		}

		// update the step action with the loaded action
		s.Action = *ca

		log.Info(fmt.Sprintf("Download action repository '%s'", s.Step.Uses))

		if s.Action.Meta.Runs.Using == model.ActionRunsUsingDocker {
			var (
				image        = ca.Meta.Runs.Image
				workspace    = ctx.Github.Workspace
				workspaceDir = ctx.Dagger.Client.Host().Directory(workspace)
			)

			switch {
			case image == "Dockerfile":
				s.container = ctx.Dagger.Client.Container().Build(ctx.Dagger.Client.Host().Directory(s.Action.Path))
			case strings.HasPrefix(image, "docker://"):
				s.container = ctx.Dagger.Client.Container().From(strings.TrimPrefix(image, "docker://"))
			default:
				// This should never happen. Adding it for safety.
				return model.ConclusionFailure, fmt.Errorf("invalid docker image: %s", image)
			}

			// add repository to the container
			s.container = s.container.WithMountedDirectory(workspace, workspaceDir).WithWorkdir(workspace)
		}

		return model.ConclusionSuccess, nil
	}
}

func (s *StepAction) preRun(stage model.StepStage) task.PreRunFn[context.Context] {
	return func(ctx *context.Context) error {
		ctx.SetAction(&s.Action)

		return ctx.SetStep(
			&model.StepRun{
				Step:    s.Step,
				Stage:   stage,
				Outputs: make(map[string]string),
				State:   make(map[string]string),
			},
		)
	}
}

func (s *StepAction) postRun() task.PostRunFn[context.Context] {
	return func(ctx *context.Context, result task.Result) {
		ctx.UnsetStep(context.RunResult(result))
		ctx.UnsetAction()
	}
}

func (s *StepAction) preCondition() task.ConditionalFn[context.Context] {
	return func(ctx *context.Context) (bool, model.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PreCondition()
		if !run {
			return false, "", nil
		}

		return evalCondition(condition, ctx)
	}
}

func (s *StepAction) pre() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case model.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PreEntrypoint)
		case model.ActionRunsUsingNode12, model.ActionRunsUsingNode16, model.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Pre)
		default:
			return model.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}

func (s *StepAction) condition() task.ConditionalFn[context.Context] {
	return func(ctx *context.Context) (bool, model.Conclusion, error) {
		return evalCondition(s.Step.If, ctx)
	}
}

func (s *StepAction) main() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case model.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.Entrypoint)
		case model.ActionRunsUsingNode12, model.ActionRunsUsingNode16, model.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Main)
		default:
			return model.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}

func (s *StepAction) postCondition() task.ConditionalFn[context.Context] {
	return func(ctx *context.Context) (bool, model.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PostCondition()
		if !run {
			return false, "", nil
		}

		return evalCondition(condition, ctx)
	}
}

func (s *StepAction) post() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		var executor Executor

		switch s.Action.Meta.Runs.Using {
		case model.ActionRunsUsingDocker:
			executor = NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PostEntrypoint)
		case model.ActionRunsUsingNode12, model.ActionRunsUsingNode16, model.ActionRunsUsingNode20:
			executor = NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Post)
		default:
			return model.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		// execute the step
		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}
