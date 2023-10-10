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
	StepRun     *core.StepRun     // Step is the current step that is being executed.
}

// SetToken sets the Github API token in the context.
func (c *Context) SetToken(token string) {
	c.Secrets.setToken(token)
	c.Github.setToken(token)
}

// SetWorkflow creates a new execution context with the given workflow and sets it to the context.
func (c *Context) SetWorkflow(wr *core.WorkflowRun) error {
	// set the workflow run to the execution context
	c.Execution = ExecutionContext{WorkflowRun: wr}

	// set the workflow run info to the github context
	c.Github.setWorkflow(wr)

	// set workflow conclusion to success explicitly
	c.Execution.WorkflowRun.Conclusion = core.ConclusionSuccess

	// set env context
	c.Env = wr.Workflow.Env

	return nil
}

// SetJob sets the given job to the execution context.
func (c *Context) SetJob(jr *core.JobRun) error {
	if c.Execution.WorkflowRun == nil {
		return errors.New("no workflow is set")
	}

	// set the job run to the execution context
	c.Execution.JobRun = jr
	c.Execution.WorkflowRun.Jobs[jr.Job.ID] = *jr

	// set the job run to the github context
	c.Github.Job = jr.Job.ID

	// set env context
	for k, v := range jr.Job.Env {
		c.Env[k] = v
	}

	// set matrix context if matrix has any values
	if len(jr.Matrix) > 0 {
		c.Matrix = jr.Matrix
	}

	// load the job context
	if err := c.LoadJob(c.Execution.WorkflowRun.Conclusion); err != nil {
		return err
	}

	// load the steps context
	if err := c.LoadSteps(); err != nil {
		return err
	}

	var needs []core.JobRun

	if len(jr.Job.Needs) > 0 {
		for _, need := range jr.Job.Needs {
			needs = append(needs, c.Execution.WorkflowRun.Jobs[need])
		}
	}

	return c.LoadNeeds(needs...)
}

// UnsetJob unsets the job from the execution context.
func (c *Context) UnsetJob() {
	jr := c.Execution.JobRun

	// update the job run in the workflow run
	c.Execution.WorkflowRun.Jobs[jr.Job.ID] = *jr

	// update workflow conclusion
	if c.Execution.WorkflowRun.Conclusion == core.ConclusionSuccess && jr.Conclusion != core.ConclusionSuccess {
		c.Execution.WorkflowRun.Conclusion = jr.Conclusion
	}

	// unset the job run from the execution context
	c.Execution.JobRun = nil

	// unset the job run from the github context
	c.Github.Job = ""

	// unset the job from env context -- just set it to the workflow env would be enough
	c.Env = c.Execution.WorkflowRun.Workflow.Env

	// reset matrix context
	c.Matrix = make(core.MatrixCombination)
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

// SetJobResults sets the status of the job.
func (c *Context) SetJobResults(conclusion, outcome core.Conclusion, outputs map[string]string) error {
	if c.Execution.JobRun == nil {
		return errors.New("no job is set")
	}

	// update current job run
	c.Execution.JobRun.Conclusion = conclusion
	c.Execution.JobRun.Outcome = outcome
	c.Execution.JobRun.Outputs = outputs

	// update job context
	c.Job.Status = conclusion

	return nil
}

// SetStep sets the given step to the execution context.
func (c *Context) SetStep(sr *core.StepRun) error {
	if c.Execution.JobRun == nil {
		return errors.New("no job is set")
	}

	c.Execution.StepRun = sr

	// set the step env context
	for k, v := range sr.Step.Environment {
		c.Env[k] = v
	}

	return nil
}

// UnsetStep unsets the step from the execution context.
func (c *Context) UnsetStep() error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	// TODO: improve this logic

	// unset the step run from the execution context

	c.Env = c.Execution.WorkflowRun.Workflow.Env

	for k, v := range c.Execution.JobRun.Job.Env {
		c.Env[k] = v
	}

	sr := c.Execution.StepRun

	// update the step run in the job run
	c.Execution.JobRun.Steps = append(c.Execution.JobRun.Steps, *sr)

	sc, ok := c.Steps[sr.Step.ID]
	if !ok {
		sc = StepContext{}
	}

	// TODO: double check this, different step stages might have different update logic for the step context
	sc.State = sr.State
	sc.Summary = sr.Summary
	sc.Outputs = sr.Outputs
	sc.Outcome = sr.Outcome
	sc.Conclusion = sr.Conclusion

	c.Steps[sr.Step.ID] = sc

	c.Execution.StepRun = nil

	return nil
}

func (c *Context) SetStepResults(conclusion, outcome core.Conclusion) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.Conclusion = conclusion
	c.Execution.StepRun.Outcome = outcome

	return nil
}

// SetStepOutput sets the output of the given step.
func (c *Context) SetStepOutput(key, value string) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.Outputs[key] = value

	return nil
}

// SetStepSummary sets the summary of the given step.
func (c *Context) SetStepSummary(summary string) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.Summary = summary

	return nil
}

// SetStepState sets the state of the given step.
func (c *Context) SetStepState(key, value string) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.State[key] = value

	return nil
}
