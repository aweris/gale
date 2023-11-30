package main

import (
	"context"
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/model"
)

type Workflows struct{}

// List returns a list of workflows and their jobs with the given options.
func (w *Workflows) List(ctx context.Context, source *Directory, workflowsDir string) ([]Workflow, error) {
	var workflows []Workflow

	walkFn := func(ctx context.Context, path string, file *File) (bool, error) {
		workflow, err := w.loadWorkflow(ctx, path, file)
		if err != nil {
			return false, err
		}

		workflows = append(workflows, *workflow)

		return true, nil
	}

	err := walkWorkflowDir(ctx, source, workflowsDir, walkFn)
	if err != nil {
		return nil, err
	}

	return workflows, nil
}

// Get returns a workflow.
func (w *Workflows) Get(ctx context.Context, source *Directory, workflow string, workflowsDir string) (*Workflow,
	error) {
	workflows, err := w.List(ctx, source, workflowsDir)
	if err != nil {
		return nil, err
	}

	for _, wf := range workflows {
		if wf.Name == workflow || wf.Path == workflow {
			return &wf, nil
		}
	}

	return nil, fmt.Errorf("workflow %s not found", workflow)
}

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

func (w *Workflows) loadWorkflow(ctx context.Context, path string, workflow *File) (*Workflow, error) {
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

type Job struct {
	// ID of the job.
	JobID string

	// Name of the job.
	Name string

	// Conditional expression to run the job.
	Condition string

	// List of jobs that must be completed before this job will run.
	Needs []string

	// Environment variables used in the job. Format: KEY=VALUE.
	Env []string

	// List of outputs of the job.
	Outputs []string

	// List of steps in the job.
	Steps []Step
}

func loadJob(id string, jm model.Job) Job {
	steps := make([]Step, len(jm.Steps))

	for i, step := range jm.Steps {
		steps[i] = loadStep(step)

		if steps[i].StepID == "" {
			steps[i].StepID = strconv.Itoa(i)
		}
	}

	name := jm.Name
	if name == "" {
		name = id
	}

	return Job{
		JobID:     id,
		Name:      name,
		Condition: jm.If,
		Needs:     jm.Needs,
		Env:       mapToKV(jm.Env),
		Steps:     steps,
	}
}

type Step struct {
	// Unique identifier of the step. Defaults to the step index in the job.
	StepID string

	// Conditional expression to run the step.
	Condition string

	// Name of the step.
	Name string

	// Action to run for the step.
	Uses string

	// Environment variables used in the step. Format: KEY=VALUE.
	Env []string

	// Inputs used in the step. Format: KEY=VALUE.
	With []string

	// Command to run for the step.
	Run string

	// Shell to use for the step.
	Shell string

	// Working directory for the step.
	WorkingDirectory string

	// Flag to continue on error.
	ContinueOnError bool

	// Maximum number of minutes to run the step.
	TimeoutMinutes int
}

func loadStep(step model.Step) Step {
	return Step{
		StepID:           step.ID,
		Condition:        step.If,
		Name:             step.Name,
		Uses:             step.Uses,
		Env:              mapToKV(step.Environment),
		With:             mapToKV(step.With),
		Run:              step.Run,
		Shell:            step.Shell,
		WorkingDirectory: step.WorkingDirectory,
		ContinueOnError:  step.ContinueOnError,
		TimeoutMinutes:   step.TimeoutMinutes,
	}
}
