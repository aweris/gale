package core

import "gopkg.in/yaml.v3"

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	ID      string            `yaml:"id"`      // ID is the ID of the job
	If      string            `yaml:"if"`      // If is the conditional expression to run the job.
	Name    string            `yaml:"name"`    // Name is the name of the job
	Needs   Needs             `yaml:"needs"`   // Needs is the list of jobs that must be completed before this job will run
	Env     map[string]string `yaml:"env"`     // Env is the environment variables used in the workflow
	Outputs map[string]string `yaml:"outputs"` // Outputs is the list of outputs of the job
	Steps   []Step            `yaml:"steps"`   // Steps is the list of steps in the job

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

// JobRun represents a single job run in a GitHub Actions workflow run
type JobRun struct {
	RunID      string            `json:"run_id"`     // RunID is the ID of the run
	Job        Job               `json:"job"`        // Job is the job to run
	Conclusion Conclusion        `json:"conclusion"` // Conclusion is the result of a completed job after continue-on-error is applied
	Outcome    Conclusion        `json:"outcome"`    // Outcome is  the result of a completed job before continue-on-error is applied
	Outputs    map[string]string `json:"outputs"`    // Outputs is the outputs generated by the job
	Steps      []StepRun         `json:"steps"`      // Steps is the list of steps in the job
}
