package core

// Workflow represents a GitHub Actions workflow.
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
type Workflow struct {
	Path string            `yaml:"-"`    // Path is the relative path to the workflow file.
	Name string            `yaml:"name"` // Name is the name of the workflow.
	Env  map[string]string `yaml:"env"`  // Env is the environment variables used in the workflow
	Jobs map[string]Job    `yaml:"jobs"` // Jobs is the list of jobs in the workflow.

	// TBD: add more fields when needed
}

type WorkflowRun struct {
	RunID         string            `json:"run_id"`         // RunID is the ID of the run
	RunNumber     string            `json:"run_number"`     // RunNumber is the number of the run
	RunAttempt    string            `json:"run_attempt"`    // RunAttempt is the attempt number of the run
	RetentionDays string            `json:"retention_days"` // RetentionDays is the number of days to keep the run logs
	Workflow      Workflow          `json:"workflow"`       // Workflow is the workflow to run
	Conclusion    Conclusion        `json:"conclusion"`     // Conclusion is the result of a completed workflow run after continue-on-error is applied
	Jobs          map[string]JobRun `json:"jobs"`           // Jobs is map of the job run id to its result
}
