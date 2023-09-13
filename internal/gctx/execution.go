package gctx

import (
	"errors"
	"path/filepath"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
)

type ExecutionContext struct {
	WorkflowRun *core.WorkflowRun // Workflow is the current workflow that is being executed.
	JobRun      *core.JobRun      // Job is the current job that is being executed.
	Step        *core.Step        // Step is the current step that is being executed.
}

// SetWorkflow creates a new execution context with the given workflow and sets it to the context.
func (c *Context) SetWorkflow(wr *core.WorkflowRun) error {
	c.Execution = ExecutionContext{WorkflowRun: wr}

	return nil
}

// UnsetWorkflow unsets the workflow from the execution context.
func (c *Context) UnsetWorkflow() {
	c.Execution.WorkflowRun = nil
}

// SetJob sets the given job to the execution context.
func (c *Context) SetJob(jr *core.JobRun) error {
	if c.Execution.WorkflowRun == nil {
		return errors.New("no workflow is set")
	}

	c.Execution.JobRun = jr

	return nil
}

// UnsetJob unsets the job from the execution context.
func (c *Context) UnsetJob() {
	c.Execution.JobRun = nil
}

// GetJobRunPath returns the path of the current job run path. If the path does not exist, it creates it.
func (c *Context) GetJobRunPath() (string, error) {
	if c.Execution.JobRun == nil {
		return "", errors.New("no job is set")
	}

	path := filepath.Join(c.path, "runs", c.Execution.WorkflowRun.RunID, "jobs", c.Execution.JobRun.RunID)

	if err := fs.EnsureDir(path); err != nil {
		return "", err
	}

	return path, nil
}

// SetStep sets the given step to the execution context.
func (c *Context) SetStep(step core.Step) error {
	if c.Execution.JobRun == nil {
		return errors.New("no job is set")
	}

	c.Execution.Step = &step

	return nil
}

// UnsetStep unsets the step from the execution context.
func (c *Context) UnsetStep() {
	c.Execution.Step = nil
}

// SetStepOutput sets the output of the given step.
func (c *Context) SetStepOutput(key, value string) error {
	if c.Execution.Step == nil {
		return errors.New("no step is set")
	}

	stepID := c.Execution.Step.ID

	sc, ok := c.Steps[stepID]
	if !ok {
		sc = StepContext{}
	}

	if sc.Outputs == nil {
		sc.Outputs = make(map[string]string)
	}

	sc.Outputs[key] = value

	c.Steps[stepID] = sc

	return nil
}

// SetStepSummary sets the summary of the given step.
func (c *Context) SetStepSummary(summary string) error {
	if c.Execution.Step == nil {
		return errors.New("no step is set")
	}

	stepID := c.Execution.Step.ID

	sc, ok := c.Steps[stepID]
	if !ok {
		sc = StepContext{}
	}

	sc.Summary = summary

	c.Steps[stepID] = sc

	return nil
}

// SetStepState sets the state of the given step.
func (c *Context) SetStepState(key, value string) error {
	if c.Execution.Step == nil {
		return errors.New("no step is set")
	}

	stepID := c.Execution.Step.ID

	sc, ok := c.Steps[stepID]
	if !ok {
		sc = StepContext{}
	}

	if sc.State == nil {
		sc.State = make(map[string]string)
	}

	sc.State[key] = value

	c.Steps[stepID] = sc

	return nil
}
