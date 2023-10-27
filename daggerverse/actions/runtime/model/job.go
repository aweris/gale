package model

import "gopkg.in/yaml.v3"

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	ID       string            `yaml:"id"`       // ID is the ID of the job
	If       string            `yaml:"if"`       // If is the conditional expression to run the job.
	Name     string            `yaml:"name"`     // Name is the name of the job
	Needs    Needs             `yaml:"needs"`    // Needs is the list of jobs that must be completed before this job will run
	Strategy Strategy          `yaml:"strategy"` // Strategy is the matrix strategy lets you use variables in a single job definition to automatically create multiple job runs that are based on the combinations of the variables.
	Env      map[string]string `yaml:"env"`      // Env is the environment variables used in the workflow
	Outputs  map[string]string `yaml:"outputs"`  // Outputs is the list of outputs of the job
	Steps    []Step            `yaml:"steps"`    // Steps is the list of steps in the job

	// TBD: add more fields when needed
}

// Needs is the list of jobs that must be completed before this job will run
type Needs []string

// UnmarshalYAML implements yaml.Unmarshaler interface for Needs. It supports both scalar and sequence nodes.
//
// Example:
//
//	needs: build # scalar node
//	needs: # sequence node
//	  - build
//	  - test
func (n *Needs) UnmarshalYAML(value *yaml.Node) error {
	var needs []string

	switch value.Kind {
	case yaml.ScalarNode:
		needs = append(needs, value.Value)
	case yaml.SequenceNode:
		for _, node := range value.Content {
			needs = append(needs, node.Value)
		}
	}

	*n = needs

	return nil
}

// Strategy represents a matrix strategy lets you use variables in a single job definition to automatically create
// multiple job runs that are based on the combinations of the variables.
type Strategy struct {
	Matrix      Matrix `yaml:"matrix"`       // Matrix is the matrix of different OS versions and other parameters
	FailFast    bool   `yaml:"fail-fast"`    // FailFast is a boolean to indicate if the job should fail immediately when a job fails.
	MaxParallel int    `yaml:"max-parallel"` // MaxParallel is the maximum number of jobs to run at a time.
}
