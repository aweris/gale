package runner

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/tools/ghx/actions"
	"github.com/aweris/gale/tools/ghx/log"
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
	default:
		return nil, fmt.Errorf("unknown step type: %s", s.Type())
	}
}

var _ Step = new(StepAction)

// StepAction is a step that runs an action.
type StepAction struct {
	runner *Runner
	ac     *actions.ExprContext
	Step   core.Step
	Action core.CustomAction
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

		return core.ConclusionSuccess, nil
	}
}

func (s *StepAction) preCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		if s.Action.Meta.Runs.Pre == "" {
			return false, "", nil
		}

		return evalStepCondition(s.Action.Meta.Runs.PreIf, s.ac)
	}
}

func (s *StepAction) pre() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionFailure, fmt.Errorf("pre run is not implemented yet")
	}
}

func (s *StepAction) mainCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		return evalStepCondition(s.Step.If, s.ac)
	}
}

func (s *StepAction) main() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionFailure, fmt.Errorf("main run is not implemented yet")
	}
}

func (s *StepAction) postCondition() TaskConditionalFn {
	return func(ctx context.Context) (bool, core.Conclusion, error) {
		if s.Action.Meta.Runs.Post == "" {
			return false, "", nil
		}

		return evalStepCondition(s.Action.Meta.Runs.PostIf, s.ac)
	}
}

func (s *StepAction) post() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionFailure, fmt.Errorf("post run is not implemented yet")
	}
}

var _ Step = new(StepRun)

// StepRun is a step that runs a job.
type StepRun struct {
	runner *Runner
	ac     *actions.ExprContext
	Step   core.Step
	Path   string // Path is the script path to run.

}

func (s *StepRun) setup() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		path := filepath.Join(config.GhxRunDir(s.runner.jr.RunID), "scripts", s.Step.ID, "run.sh")

		content := []byte(fmt.Sprintf("#!/bin/bash\n%s", s.Step.Run))

		err := fs.WriteFile(path, content, 0755)
		if err != nil {
			return core.ConclusionFailure, err
		}

		s.Path = path

		// make it debug level because it's not really important and it's visible in Github Actions logs
		log.Debug(fmt.Sprintf("Write script to '%s' for step '%s'", path, s.Step.ID))

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
		return evalStepCondition(s.Step.If, s.ac)
	}
}

func (s *StepRun) main() TaskExecutorFn {
	return func(ctx context.Context) (core.Conclusion, error) {
		return core.ConclusionFailure, fmt.Errorf("main run is not implemented yet")
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
