package actions

// Jobs represents a map of jobs
// For more information about workflows, see: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobs
type Jobs map[string]*Job

// Job represents a single job in a GitHub Actions workflow
// For more information about workflows, see: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	// Name is the name of the job
	Name string `yaml:"name"`

	// Environment is the environment variables used in the job
	Environment Environment `yaml:"env"`

	// Steps is a list of steps
	Steps Steps `yaml:"steps"`
}
