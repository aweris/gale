package core

// Workflow represents a GitHub Actions workflow.
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions
type Workflow struct {
	Path string         `yaml:"-"`    // Path is the relative path to the workflow file.
	Name string         `yaml:"name"` // Name is the name of the workflow.
	Jobs map[string]Job `yaml:"jobs"` // Jobs is the list of jobs in the workflow.

	// TBD: add more fields when needed
}

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	Name string `yaml:"name"` // Name is the name of the job

	// TBD: add more fields when needed
}
