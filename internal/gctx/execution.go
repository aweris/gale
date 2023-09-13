package gctx

import (
	"errors"

	"github.com/aweris/gale/internal/core"
)

type ExecutionContext struct {
	Workflow *core.Workflow // Workflow is the current workflow that is being executed.
	Job      *core.Job      // Job is the current job that is being executed.
	Step     *core.Step     // Step is the current step that is being executed.
}

// SetWorkflow creates a new execution context with the given workflow and sets it to the context.
func (c *Context) SetWorkflow(wf core.Workflow) error {
	c.Execution = ExecutionContext{Workflow: &wf}

	return nil
}

// UnsetWorkflow unsets the workflow from the execution context.
func (c *Context) UnsetWorkflow() {
	c.Execution.Workflow = nil
}

// SetJob sets the given job to the execution context.
func (c *Context) SetJob(job core.Job) error {
	if c.Execution.Workflow == nil {
		return errors.New("no workflow is set")
	}

	c.Execution.Job = &job

	return nil
}

// UnsetJob unsets the job from the execution context.
func (c *Context) UnsetJob() {
	c.Execution.Job = nil
}

// SetStep sets the given step to the execution context.
func (c *Context) SetStep(step core.Step) error {
	if c.Execution.Job == nil {
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
