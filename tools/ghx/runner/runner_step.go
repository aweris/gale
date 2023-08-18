package runner

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/log"
	"github.com/aweris/gale/tools/ghx/actions"
)

type Step interface {
	// setup returns the function that sets up the step before execution.
	setup() TaskExecutorFn

	// preCondition returns the function that checks if the pre execution condition is met.
	preCondition() TaskConditionalFn

	// pre returns the function that executes the pre execution logic just before the main execution.
	pre() TaskExecutorFn

	// mainCondition returns the function that checks if the main execution condition is met.
	mainCondition() TaskConditionalFn

	// main returns the function that executes the main execution logic.
	main() TaskExecutorFn

	// postCondition returns the function that checks if the post execution condition is met.
	postCondition() TaskConditionalFn

	// post returns the function that executes the post execution logic just after the main execution.
	post() TaskExecutorFn
}

// NewStep creates a new step from the given step configuration.
func NewStep(runner *Runner, s core.Step) (Step, error) {
	switch s.Type() {
	case core.StepTypeAction:
		return &StepAction{runner: runner, Step: s}, nil
	case core.StepTypeRun:
		return &StepRun{runner: runner, Step: s}, nil
	case core.StepTypeDocker:
		return &StepDocker{runner: runner, Step: s}, nil
	default:
		return nil, fmt.Errorf("unknown step type: %s", s.Type())
	}
}

var _ Step = new(StepAction)

// StepAction is a step that runs an action.
type StepAction struct {
	runner    *Runner
	container *dagger.Container
	Step      core.Step
	Action    core.CustomAction
}

func (s *StepAction) setup() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		ca, err := actions.LoadActionFromSource(ctx, s.Step.Uses)
		if err != nil {
			return core.ConclusionFailure, err
		}

		// update the step action with the loaded action
		s.Action = *ca

		log.Info(fmt.Sprintf("Download action repository '%s'", s.Step.Uses))

		if s.Action.Meta.Runs.Using == core.ActionRunsUsingDocker {
			var (
				image        = ca.Meta.Runs.Image
				workspace    = s.runner.context.Github.Workspace
				workspaceDir = config.Client().Host().Directory(workspace)
			)

			switch {
			case image == "Dockerfile":
				s.container = config.Client().Container().Build(ca.Dir)
			case strings.HasPrefix(image, "docker://"):
				s.container = config.Client().Container().From(strings.TrimPrefix(image, "docker://"))
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

func (s *StepAction) preCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PreCondition()
		if !run {
			return false, "", nil
		}

		return evalStepCondition(condition, s.runner.context)
	}
}

func (s *StepAction) pre() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			cmd := NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PreEntrypoint)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			cmd := NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Pre)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) mainCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, s.runner.context)
	}
}

func (s *StepAction) main() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			cmd := NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.Entrypoint)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			cmd := NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Main)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) postCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		run, condition := s.Action.Meta.Runs.PostCondition()
		if !run {
			return false, "", nil
		}

		return evalStepCondition(condition, s.runner.context)
	}
}

func (s *StepAction) post() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		switch s.Action.Meta.Runs.Using {
		case core.ActionRunsUsingDocker:
			cmd := NewContainerExecutorFromStepAction(s, s.Action.Meta.Runs.PostEntrypoint)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		case core.ActionRunsUsingNode12, core.ActionRunsUsingNode16:
			cmd := NewCmdExecutorFromStepAction(s, s.Action.Meta.Runs.Post)

			err := cmd.Execute(ctx)
			if err != nil && !s.Step.ContinueOnError {
				return core.ConclusionFailure, err
			}
		default:
			return core.ConclusionFailure, fmt.Errorf("invalid action runs using: %s", s.Action.Meta.Runs.Using)
		}

		return core.ConclusionSuccess, nil
	}
}

var _ Step = new(StepRun)

// StepRun is a step that runs a job.
type StepRun struct {
	runner    *Runner
	Step      core.Step
	Shell     string   // Shell is the shell to use to run the script.
	ShellArgs []string // ShellArgs are the arguments to pass to the shell.
	Path      string   // Path is the script path to run.

}

func (s *StepRun) setup() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionSuccess, nil
	}
}

// preCondition returns always false because pre run is not supported for StepRun.
func (s *StepRun) preCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return false, "", nil
	}
}

func (s *StepRun) pre() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionSkipped, nil
	}
}

func (s *StepRun) mainCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, s.runner.context)
	}
}

func (s *StepRun) main() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		path := filepath.Join(config.GhxRunDir(s.runner.jr.RunID), "scripts", s.Step.ID, "run.sh")

		// evaluate run script against the expressions
		run, err := evalString(s.Step.Run, s.runner.context)
		if err != nil {
			return core.ConclusionFailure, err
		}

		content := []byte(fmt.Sprintf("#!/bin/bash\n%s", run))

		err = fs.WriteFile(path, content, 0755)
		if err != nil {
			return core.ConclusionFailure, err
		}

		s.Path = path
		s.Shell = "bash"
		s.ShellArgs = []string{"--noprofile", "--norc", "-e", "-o", "pipefail"}

		cmd := NewCmdExecutorFromStepRun(s)

		err = cmd.Execute(ctx)
		if err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

// postCondition returns always false because post run is not supported for StepRun.
func (s *StepRun) postCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return false, "", nil
	}
}

func (s *StepRun) post() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionSkipped, nil
	}
}

var _ Step = new(StepDocker)

type StepDocker struct {
	runner    *Runner
	container *dagger.Container
	Step      core.Step
}

func (s *StepDocker) setup() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		var (
			image        = strings.TrimPrefix(s.Step.Uses, "docker://")
			workspace    = s.runner.context.Github.Workspace
			workspaceDir = config.Client().Host().Directory(workspace)
		)

		// configure the step container
		s.container = config.Client().
			Container().
			From(image).
			WithMountedDirectory(workspace, workspaceDir).
			WithWorkdir(workspace)

		// TODO: This will be print same log line if the image used multiple times. However, this scenario is not really common and no benefit to fix this scenario for now.
		log.Info(fmt.Sprintf("Pull '%s'", image))

		return core.ConclusionSuccess, nil
	}
}

// preCondition returns always false because pre run is not supported for StepDocker.
func (s *StepDocker) preCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return false, "", nil
	}
}

func (s *StepDocker) pre() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionSkipped, nil
	}
}

func (s *StepDocker) mainCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, s.runner.context)
	}
}

func (s *StepDocker) main() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		executor := NewContainerExecutorFromStepDocker(s)

		err := executor.Execute(ctx)
		if err != nil && !s.Step.ContinueOnError {
			return core.ConclusionFailure, err
		}

		return core.ConclusionSuccess, nil
	}
}

// postCondition returns always false because post run is not supported for StepDocker.
func (s *StepDocker) postCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return false, "", nil
	}
}

func (s *StepDocker) post() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionSkipped, nil
	}
}
