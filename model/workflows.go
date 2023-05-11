package model

import "dagger.io/dagger"

// Workflows represents a collection of workflow.
type Workflows map[string]*Workflow

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	Path string       `yaml:"-"` // path is the relative path to the workflow file.
	File *dagger.File `yaml:"-"` // File is the raw content of the workflow file.

	Name        string            `yaml:"name"` // Name is the name of the workflow
	Environment map[string]string `yaml:"env"`  // Environment is the environment variables used in the workflow
	Jobs        Jobs              `yaml:"jobs"` // Jobs is the list of jobs in the workflow.

	// TBD -- we'll add more fields here as we need them.
}
