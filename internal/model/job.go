package model

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	Name  string            `yaml:"name"`  // Name is the name of the job
	Env   map[string]string `yaml:"env"`   // Environment is the environment variables used in the workflow
	Steps []Step            `yaml:"steps"` // Steps is a list of steps

	// TBD -- we'll add more fields here as we need them.
}
