package actions

// Steps represents a list of steps
type Steps []*Step

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
	Environment Environment `yaml:"env,omitempty"`

	// With maps input names to their values for the step.
	With map[string]string `yaml:"with,omitempty"`
}

// StepStatus represents the status of a step
type StepStatus string

const (
	StepStatusSuccess StepStatus = "success"
	StepStatusFailure StepStatus = "failure"
	StepStatusSkipped StepStatus = "skipped"
)

type StepResult struct {
	Outputs    map[string]string
	Conclusion StepStatus
	Outcome    StepStatus
}
