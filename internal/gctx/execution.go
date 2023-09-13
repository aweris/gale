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
