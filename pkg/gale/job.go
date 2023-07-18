package gale

import (
	"context"
	"encoding/json"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/google/uuid"

	"github.com/aweris/gale/internal/model"
	"github.com/aweris/gale/pkg/repository"
)

var _ With = new(Job)

// Job is the configuration for a job to execute with gale.
type Job struct {
	client *dagger.Client   // dagger client
	config *model.JobConfig // configuration of the job to execute

	results *dagger.Directory // results directory where all job related files and configurations are stored
}

func (g *Gale) Job() *Job {
	return &Job{
		client: g.client,
		config: &model.JobConfig{},
	}
}

func (j *Job) Load(ctx context.Context, workflow, job string) (*Job, error) {
	// load all workflows
	workflows, err := repository.LoadWorkflows(ctx, j.client)
	if err != nil {
		return j, err
	}

	// load workflow
	wf, ok := workflows[workflow]
	if !ok {
		return j, ErrWorkflowNotFound
	}

	// load job
	jm, ok := wf.Jobs[job]
	if !ok {
		return j, ErrJobNotFound
	}

	// set job to job config
	j.config.Job = jm

	// TODO: add rest of the required fields to job config

	return j, nil
}

func (j *Job) WithContainerFunc(container *dagger.Container) *dagger.Container {
	data, err := json.Marshal(j.config)
	if err != nil {
		fail(container, err)
	}

	// TODO: look better way to generate run id, uuid is unique but not readable

	runID := uuid.New().String()

	// unique path for the job
	path := filepath.Join(containerRunnerPath, "runs", runID)

	// create initial directory for the job and mount it to the container
	dir := j.client.Directory().WithNewFile("config.json", string(data))

	container = container.WithMountedDirectory(path, dir)
	container = container.WithExec([]string{"ghx", "run", runID})

	// load final directory for the job
	j.results = container.Directory(path)

	return container
}
