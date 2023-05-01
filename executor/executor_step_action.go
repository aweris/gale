package executor

import (
	"context"
	"fmt"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/logger"
	runnerpkg "github.com/aweris/gale/runner"
)

var _ StepExecutor = new(StepActionExecutor)

// StepActionExecutor represents a step executor that executes an custom action.
type StepActionExecutor struct {
	// step is the information about the step.
	step *gha.Step

	// log is the logger used by the executor.
	log logger.Logger

	// fallbackEnvs is the list of environment variables that should use the fallback value when removing the step
	// environment variables.
	fallbackEnvs []gha.Environment
}

// NewStepActionExecutor creates a new step action executor a
func NewStepActionExecutor(step *gha.Step, log logger.Logger, fallbackEnvs ...gha.Environment) *StepActionExecutor {
	// TODO: implement docker and run step executors as well. Currently only action is implemented.
	return &StepActionExecutor{
		step:         step,
		log:          log,
		fallbackEnvs: fallbackEnvs,
	}
}

func (s *StepActionExecutor) pre(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		step   = s.step
		path   = runner.ActionPathsBySource[step.Uses]
		action = runner.ActionsBySource[step.Uses]
	)

	// TODO: check pre-if as well
	if action.Runs.Pre == "" {
		return nil
	}

	s.log.Info(fmt.Sprintf("Pre Run %s", step.Uses))

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	_, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Pre))
	if outErr != nil {
		return outErr
	}

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}

func (s *StepActionExecutor) main(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		step   = s.step
		path   = runner.ActionPathsBySource[step.Uses]
		action = runner.ActionsBySource[step.Uses]
	)

	s.log.Info(fmt.Sprintf("Run %s", step.Uses))

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	_, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Main))
	if outErr != nil {
		return outErr
	}

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}

func (s *StepActionExecutor) post(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		step   = s.step
		path   = runner.ActionPathsBySource[step.Uses]
		action = runner.ActionsBySource[step.Uses]
	)

	// TODO: check post-if as well
	if action.Runs.Post == "" {
		return nil
	}

	s.log.Info(fmt.Sprintf("Post Run %s", step.Uses))

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	_, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Post))
	if outErr != nil {
		return outErr
	}

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}
