package model

// Workflows represents a collection of workflow.
type Workflows map[string]*Workflow

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	Path string                 `yaml:"-"`    // path is the relative path to the workflow file.
	Name string                 `yaml:"name"` // Name is the name of the workflow
	Jobs map[string]interface{} `yaml:"jobs"` // Jobs is the list of jobs in the workflow.
}
