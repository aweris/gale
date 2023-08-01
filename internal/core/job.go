package core

import (
	"context"
	"encoding/json"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
)

// Job represents a single job in a GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_id
type Job struct {
	Name  string            `yaml:"name"`  // Name is the name of the job
	Env   map[string]string `yaml:"env"`   // Env is the environment variables used in the workflow
	Steps []Step            `yaml:"steps"` // Steps is the list of steps in the job

	// TBD: add more fields when needed
}

type JobRun struct {
	RunID string `json:"runID"` // RunID is the ID of the run
	Job   Job    `json:"job"`   // Job is the job to run
}

// MarshalJobRunToDir marshals the job run to a dagger directory
func MarshalJobRunToDir(_ context.Context, jobRun *JobRun) (*dagger.Directory, error) {
	data, err := json.Marshal(jobRun)
	if err != nil {
		return nil, err
	}

	dir := config.Client().Directory()

	dir = dir.WithNewFile("job_run.json", string(data))

	return dir, nil
}

// UnmarshalJobRunFromDir unmarshal the job run from a dagger directory
func UnmarshalJobRunFromDir(ctx context.Context, dir *dagger.Directory) (*JobRun, error) {
	jr := &JobRun{}

	data, err := dir.File("job_run.json").Contents(ctx)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(data), jr); err != nil {
		return nil, err
	}

	return jr, nil
}
