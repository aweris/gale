package model

// Step represents a single task in a job context at GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	ID               string            `yaml:"id,omitempty"`                // ID is the unique identifier of the step.
	If               string            `yaml:"if,omitempty"`                // If is the conditional expression to run the step.
	Name             string            `yaml:"name,omitempty"`              // Name is the name of the step.
	Uses             string            `yaml:"uses,omitempty"`              // Uses is the action to run for the step.
	Environment      map[string]string `yaml:"env,omitempty"`               // Environment maps environment variable names to their values.
	With             map[string]string `yaml:"with,omitempty"`              // With maps input names to their values for the step.
	Run              string            `yaml:"run,omitempty"`               // Run is the command to run for the step.
	Shell            string            `yaml:"shell,omitempty"`             // Shell is the shell to use for the step.
	WorkingDirectory string            `yaml:"working-directory,omitempty"` // WorkingDirectory is the working directory for the step.
	ContinueOnError  bool              `yaml:"continue-on-error,omitempty"` // ContinueOnError is a flag to continue on error.
	TimeoutMinutes   int               `yaml:"timeout-minutes,omitempty"`   // TimeoutMinutes is the maximum number of minutes to run the step.
}
