package main

import "github.com/aweris/gale/common/model"

type Step struct {
	// Unique identifier of the step. Defaults to the step index in the job.
	StepID string

	// Conditional expression to run the step.
	Condition string

	// Name of the step.
	Name string

	// Action to run for the step.
	Uses string

	// Environment variables used in the step. Format: KEY=VALUE.
	Env []string

	// Inputs used in the step. Format: KEY=VALUE.
	With []string

	// Command to run for the step.
	Run string

	// Shell to use for the step.
	Shell string

	// Working directory for the step.
	WorkingDirectory string

	// Flag to continue on error.
	ContinueOnError bool

	// Maximum number of minutes to run the step.
	TimeoutMinutes int
}

func loadStep(step model.Step) Step {
	return Step{
		StepID:           step.ID,
		Condition:        step.If,
		Name:             step.Name,
		Uses:             step.Uses,
		Env:              mapToKV(step.Environment),
		With:             mapToKV(step.With),
		Run:              step.Run,
		Shell:            step.Shell,
		WorkingDirectory: step.WorkingDirectory,
		ContinueOnError:  step.ContinueOnError,
		TimeoutMinutes:   step.TimeoutMinutes,
	}
}
