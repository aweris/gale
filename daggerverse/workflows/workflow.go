package main

import (
	"context"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/model"
)

type Workflow struct {
	// Relative path of the workflow file.
	Path string

	// Workflow file source.
	Src *File

	// Workflow name. Defaults to the file path.
	Name string

	// Environment variables used in the workflow. Format: KEY=VALUE.
	Env []string

	// Jobs in the workflow.
	Jobs []Job
}

func loadWorkflow(ctx context.Context, path string, workflow *File) (*Workflow, error) {
	var wm model.Workflow

	data, err := workflow.Contents(ctx)
	if err != nil {
		return nil, err
	}

	// TODO: refactor this when dagger supports map types. This is a temporary workaround.
	if err := yaml.Unmarshal([]byte(data), &wm); err != nil {
		return nil, err
	}

	jobs := make([]Job, 0, len(wm.Jobs))

	for id, job := range wm.Jobs {
		jobs = append(jobs, loadJob(id, job))
	}

	return &Workflow{
		Path: path,
		Src:  workflow,
		Name: wm.Name,
		Env:  mapToKV(wm.Env),
		Jobs: jobs,
	}, nil
}

// Returns the YAML representation of the workflow.
func (w *Workflow) Yaml(ctx context.Context) (string, error) {
	return w.Src.Contents(ctx)
}
