package core

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	Name string `yaml:"name"` // Name is the name of the job

	// TBD: add more fields when needed
}
