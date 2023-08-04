package core

import (
	"context"
	"encoding/json"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
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

var _ helpers.WithContainerFuncHook = new(JobRun)

// JobRun represents a job run configuration that is passed to the container
type JobRun struct {
	RunID string `json:"runID"` // RunID is the ID of the run
	Job   Job    `json:"job"`   // Job is the job to run
}

// NewJobRun creates a new job run
func NewJobRun(runID string, job Job) JobRun {
	return JobRun{
		RunID: runID,
		Job:   job,
	}
}

func (j JobRun) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		data, err := json.Marshal(j)
		if err != nil {
			return helpers.FailPipeline(container, err)
		}

		if len(data) == 0 {
			return helpers.FailPipeline(container, fmt.Errorf("job run is empty"))
		}

		dir := config.Client().Directory()

		dir = dir.WithNewFile("job_run.json", string(data))

		return container.WithDirectory(config.GhxRunDir(j.RunID), dir)
	}
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
