package model

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
