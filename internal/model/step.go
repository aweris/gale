package model

// Step represents a single task in a job context at GitHub Actions workflow
// For more information about workflows, see: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	// ID is the unique identifier of the step.
	ID string `yaml:"id,omitempty"`

	// Name is the name of the step.
	Name string `yaml:"name"`

	// Uses is the action to run for the step.
	Uses string `yaml:"uses,omitempty"`

	// Environment maps environment variable names to their values.
	Environment map[string]string `yaml:"env,omitempty"`

	// With maps input names to their values for the step.
	With map[string]string `yaml:"with,omitempty"`

	// Run is the command to run for the step.
	Run string `yaml:"run,omitempty"`

	// Shell is the shell to use for the step.
	Shell string `yaml:"shell,omitempty"`
}

// StepType represents the type of step
type StepType string

const (
	StepTypeAction  StepType = "action"
	StepTypeRun     StepType = "run"
	StepTypeUnknown StepType = "unknown"
	//TODO: add support for docker and composite steps types
)

func (s *Step) Type() StepType {
	if s.Uses != "" {
		return StepTypeAction
	}

	if s.Run != "" {
		return StepTypeRun
	}

	return StepTypeUnknown
}
