package model

// Jobs represents a map of jobs
// For more information about workflows, see: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
type Jobs map[string]*Job

// Job represents a single job in a GitHub Actions workflow
// For more information about workflows, see: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	Name        string            `yaml:"name"`  // Name is the name of the job
	Environment map[string]string `yaml:"env"`   // Environment is the environment variables used in the workflow
	Steps       Steps             `yaml:"steps"` // Steps is a list of steps
	// TBD -- we'll add more fields here as we need them.
}
