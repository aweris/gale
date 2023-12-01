package context

import (
	"errors"
	"path/filepath"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/common/model"
)

// SetWorkflow creates a new execution context with the given workflow and sets it to the context.
func (c *Context) SetWorkflow(wr *model.WorkflowRun) error {
	// set the workflow run to the execution context
	c.Execution = ExecutionContext{WorkflowRun: wr}

	// set workflow conclusion to success explicitly
	c.Execution.WorkflowRun.Conclusion = model.ConclusionSuccess

	// set env context
	c.Env = wr.Workflow.Env

	return nil
}

func (c *Context) UnsetWorkflow(result RunResult) {
	// ignoring error since directory must exist at this point of execution
	dir, _ := c.GetWorkflowRunPath()

	report := NewWorkflowRunReport(&result, c.Github, c.Execution.WorkflowRun)

	if err := fs.WriteJSONFile(filepath.Join(dir, "workflow_run.json"), report); err != nil {
		log.Errorf("failed to write workflow run", "error", err, "workflow", c.Execution.WorkflowRun.Workflow.Name)
	}
}

// SetJob sets the given job to the execution context.
func (c *Context) SetJob(jr *model.JobRun) error {
	if c.Execution.WorkflowRun == nil {
		return errors.New("no workflow is set")
	}

	// set the job run to the execution context
	c.Execution.JobRun = jr

	// set the job run to the github context
	c.Github.Job = jr.Job.ID

	// set env context
	for k, v := range jr.Job.Env {
		c.Env[k] = v
	}

	// set matrix context if matrix has any values
	if len(jr.Matrix) > 0 {
		c.Matrix = MatrixContext(jr.Matrix)
	}

	// load the job context
	c.Job = JobContext{Status: c.Execution.WorkflowRun.Conclusion}

	// load the steps context
	c.Steps = make(StepsContext)

	c.Needs = make(NeedsContext)

	// ignoring error since directory must exist at this point of execution
	dir, _ := c.GetWorkflowRunPath()

	if len(jr.Job.Needs) > 0 {
		for _, need := range jr.Job.Needs {
			path := filepath.Join(dir, "jobs", need, "job_run.json")

			var jr model.JobRun

			fs.ReadJSONFile(path, &jr)

			c.Needs[need] = NeedContext{Result: jr.Conclusion, Outputs: jr.Outputs}
		}
	}

	return nil
}

// UnsetJob unsets the job from the execution context.
func (c *Context) UnsetJob(result RunResult) {
	jr := c.Execution.JobRun

	// update workflow conclusion
	if c.Execution.WorkflowRun.Conclusion == model.ConclusionSuccess && jr.Conclusion != model.ConclusionSuccess {
		c.Execution.WorkflowRun.Conclusion = jr.Conclusion
	}
	// unset the job run from the github context
	c.Github.Job = ""

	// unset the job from env context -- just set it to the workflow env would be enough
	c.Env = c.Execution.WorkflowRun.Workflow.Env

	// reset matrix context
	c.Matrix = make(MatrixContext)

	// write the job run result to the file system
	// ignoring error since directory must be exist at this point of execution
	dir, _ := c.GetJobRunPath()

	report := NewJobRunReport(&result, c.Execution.JobRun)

	if err := fs.WriteJSONFile(filepath.Join(dir, "job_run.json"), report); err != nil {
		log.Errorf("failed to write job run", "error", err, "workflow", c.Execution.WorkflowRun.Workflow.Name)
	}

	// unset the job run from the execution context
	c.Execution.JobRun = nil
}

// SetJobResults sets the status of the job.
func (c *Context) SetJobResults(conclusion, outcome model.Conclusion, outputs map[string]string) error {
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
func (c *Context) SetStep(sr *model.StepRun) error {
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
func (c *Context) UnsetStep(result RunResult) {
	if c.Execution.StepRun == nil {
		return
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

	// only export the result of the main stage
	if c.Execution.StepRun.Stage == model.StepStageMain {
		// write the job run result to the file system
		// ignoring error since directory must exist at this point of execution
		dir, _ := c.GetStepRunPath()

		report := NewStepRunReport(&result, c.Execution.StepRun)

		if err := fs.WriteJSONFile(filepath.Join(dir, "step_run.json"), &report); err != nil {
			log.Errorf("failed to write step run", "error", err, "workflow", c.Execution.WorkflowRun.Workflow.Name)
		}

		if c.Execution.StepRun.Summary != "" {
			if err := fs.WriteFile(filepath.Join(dir, "summary.md"), []byte(c.Execution.StepRun.Summary), 0600); err != nil {
				log.Errorf("failed to write step run summary", "error", err, "workflow", c.Execution.WorkflowRun.Workflow.Name)
			}
		}
	}

	c.Execution.StepRun = nil
}

func (c *Context) SetStepResults(conclusion, outcome model.Conclusion) error {
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

// AddStepPath adds the given path to the step path.
func (c *Context) AddStepPath(path string) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.Path = append(c.Execution.StepRun.Path, path)

	return nil
}

func (c *Context) SetStepEnv(key, value string) error {
	if c.Execution.StepRun == nil {
		return errors.New("no step is set")
	}

	c.Execution.StepRun.Environment[key] = value

	return nil
}

func (c *Context) SetAction(action *model.CustomAction) {
	c.Execution.CurrentAction = action
}

func (c *Context) UnsetAction() {
	c.Execution.CurrentAction = nil
}
