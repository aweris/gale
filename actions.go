package main

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/model"
)

type Actions struct {
	// The repository information
	Repo *RepoInfo

	// The workflows in the repository
	Workflows *Workflows

	// The steps in the action
	Steps []Step
}

func (a *Actions) Action(
	// ID of the step. Defaults to the step index in the job.
	// +optional=true
	stepID string,
	// External workflow file to run.
	// +optional=false
	uses string,
	// Environment variables for the action. Format: name=value.
	// +optional=true
	env []string,
	// Input parameters for the action. Format: name=value.
	// +optional=true
	with []string,
) (*Actions, error) {
	envVal, err := ParseKeyValuePairs(env)
	if err != nil {
		return nil, err
	}

	withVal, err := ParseKeyValuePairs(with)
	if err != nil {
		return nil, err
	}

	step := Step{
		StepID: stepID,
		Uses:   uses,
		Env:    envVal,
		With:   withVal,
	}

	a.Steps = append(a.Steps, step)

	return a, nil
}

func (a *Actions) Run(
	// The action to run. it should be in the format of <action>@<version>
	ctx context.Context,
	// Name of the event that triggered the workflow. e.g. push
	// +optional=true
	// +default=push
	event string,
	// File with the complete webhook event payload.
	// +optional=true
	eventFile *File,
	// Container to use for the runner(default: ghcr.io/catthehacker/ubuntu:act-latest).
	// +optional=true
	container *Container,
	// Enables debug mode.
	// +optional=true
	// +default=false
	runnerDebug bool,
	// GitHub token to use for authentication.
	// +optional=true
	token *Secret,
) (*WorkflowRun, error) {
	models := make([]model.Step, 0, len(a.Steps))

	for _, step := range a.Steps {
		sm := model.Step{
			ID:          step.StepID,
			Uses:        step.Uses,
			Environment: ConvertKVSliceToMap(step.Env),
			With:        ConvertKVSliceToMap(step.With),
		}

		models = append(models, sm)
	}

	w := &model.Workflow{
		Jobs: map[string]model.Job{
			"dagger": {
				Steps: models,
			},
		},
	}

	data, err := yaml.Marshal(w)
	if err != nil {
		return nil, err
	}

	fmt.Printf("workflow: %s\n", string(data))

	if eventFile == nil {
		eventFile = dag.Directory().WithNewFile("event.json", "{}").File("event.json")
	}

	if container == nil {
		container = dag.Container().From("ghcr.io/catthehacker/ubuntu:act-latest")
	}

	planner := NewWorkflowExecutionPlanner(
		a.Repo,
		a.Workflows,
		&WorkflowRunOpts{WorkflowFile: dag.Directory().WithNewFile("workflow.yml", string(data)).File("workflow.yml")},
		&RunnerOpts{Ctr: container, Debug: runnerDebug},
		&EventOpts{Name: event, File: eventFile},
		&SecretOpts{Token: token},
	)

	executor, err := planner.Plan(ctx)
	if err != nil {
		return nil, err
	}

	return executor.Execute(ctx)
}
