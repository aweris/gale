package core

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	ID      string            `yaml:"id"`      // ID is the ID of the job
	Name    string            `yaml:"name"`    // Name is the name of the job
	Env     map[string]string `yaml:"env"`     // Env is the environment variables used in the workflow
	Outputs map[string]string `yaml:"outputs"` // Outputs is the list of outputs of the job
	Steps   []Step            `yaml:"steps"`   // Steps is the list of steps in the job

	// TBD: add more fields when needed
}

// JobRun represents a single job run in a GitHub Actions workflow run
type JobRun struct {
	RunID      string            `json:"run_id"`     // RunID is the ID of the run
	Job        Job               `json:"job"`        // Job is the job to run
	Conclusion Conclusion        `json:"conclusion"` // Conclusion is the result of a completed job after continue-on-error is applied
	Outcome    Conclusion        `json:"outcome"`    // Outcome is  the result of a completed job before continue-on-error is applied
	Outputs    map[string]string `json:"outputs"`    // Outputs is the outputs generated by the job
}
