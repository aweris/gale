package main

import (
	"strconv"

	"github.com/aweris/gale/common/model"
)

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
