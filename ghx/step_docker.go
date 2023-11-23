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
	_ Step      = new(StepDocker)
	_ SetupHook = new(StepDocker)
)

type StepDocker struct {
	container *dagger.Container
	Step      model.Step
}

func (s *StepDocker) setup() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		var (
			image        = strings.TrimPrefix(s.Step.Uses, "docker://")
			workspace    = ctx.Github.Workspace
			workspaceDir = ctx.Dagger.Client.Host().Directory(workspace)
		)

		// configure the step container
		s.container = ctx.Dagger.Client.
			Container().
			From(image).
			WithMountedDirectory(workspace, workspaceDir).
			WithWorkdir(workspace)

		// TODO: This will be print same log line if the image used multiple times. However, this scenario is not really common and no benefit to fix this scenario for now.
		log.Info(fmt.Sprintf("Pull '%s'", image))

		return model.ConclusionSuccess, nil
	}
}

func (s *StepDocker) condition() task.ConditionalFn[context.Context] {
	return func(ctx *context.Context) (bool, model.Conclusion, error) {
		return evalCondition(s.Step.If, ctx)
	}
}

func (s *StepDocker) main() task.RunFn[context.Context] {
	return func(ctx *context.Context) (model.Conclusion, error) {
		executor := NewContainerExecutorFromStepDocker(s)

		return executeStep(ctx, executor, s.Step.ContinueOnError)
	}
}
