package executor

import (
	"context"
	"fmt"

	"github.com/aweris/gale/gha"
	runnerpkg "github.com/aweris/gale/runner"
)

var _ StepExecutor = new(StepActionExecutor)

// StepActionExecutor represents a step executor that executes an custom action.
type StepActionExecutor struct {
	// step is the information about the step.
	step *gha.Step

	// action is the information about the action used by the step.
	action *gha.Action

	// path is the path to the action in the container.
	path string

	// fallbackEnvs is the list of environment variables that should use the fallback value when removing the step
	// environment variables.
	fallbackEnvs []gha.Environment
}

// NewStepActionExecutor creates a new step action executor a
func NewStepActionExecutor(step *gha.Step, action *gha.Action, path string, fallbackEnvs ...gha.Environment) *StepActionExecutor {
	// TODO: implement docker and run step executors as well. Currently only action is implemented.
	return &StepActionExecutor{
		step:   step,
		action: action,
		path:   path,
	}
}

func (s *StepActionExecutor) pre(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		path   = s.path
		step   = s.step
		action = s.action
	)

	// TODO: check pre-if as well
	if action.Runs.Pre == "" {
		return nil
	}

	fmt.Printf("Pre Run %s\n", step.Uses)

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	out, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Pre))
	if outErr != nil {
		return outErr
	}

	fmt.Printf("Pre step output: %s\n", out)

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}

func (s *StepActionExecutor) main(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		path   = s.path
		step   = s.step
		action = s.action
	)

	fmt.Printf("Run %s\n", step.Uses)

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	out, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Main))
	if outErr != nil {
		return outErr
	}

	fmt.Printf("Pre step output: %s\n", out)

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}

func (s *StepActionExecutor) post(ctx context.Context, runner *runnerpkg.Runner) error {
	var (
		path   = s.path
		step   = s.step
		action = s.action
	)

	// TODO: check post-if as well
	if action.Runs.Post == "" {
		return nil
	}

	fmt.Printf("Post Run %s\n", step.Uses)

	// Set up inputs and environment variables for step
	runner.WithEnvironment(step.Environment)
	runner.WithInputs(step.With)

	// Execute main step
	// TODO: add error handling. Need to check step continue-on-error, fail, always conditions as well
	out, outErr := runner.ExecAndCaptureOutput(ctx, "node", fmt.Sprintf("%s/%s", path, action.Runs.Post))
	if outErr != nil {
		return outErr
	}

	fmt.Printf("Pre step output: %s\n", out)

	// Clean up inputs and environment variables for next step
	runner.WithoutInputs(step.With)
	runner.WithoutEnvironment(step.Environment, s.fallbackEnvs...)

	return nil
}
