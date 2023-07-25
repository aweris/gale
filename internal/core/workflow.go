package core

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	Path string `yaml:"-"`    // Path is the relative path to the workflow file.
	Name string `yaml:"name"` // Name is the name of the workflow.

	// TBD: add more fields when needed
}
