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

// JobRun represents a job run configuration that is passed to the container
type JobRun struct {
	RunID string `json:"runID"` // RunID is the ID of the run
	Job   Job    `json:"job"`   // Job is the job to run
}

// NewJobRun creates a new job run
func NewJobRun(runID string, job Job) JobRun {
	return JobRun{
		RunID: runID,
		Job:   job,
	}
}
