package ghx

import (
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/gctx"
)

type ExprContext struct {
	*gctx.Context
}

// TODO: we'll remove this slowly and replace it with the new context.

func NewExprContext(ctx *gctx.Context) (*ExprContext, error) {
	return &ExprContext{ctx}, nil
}

// WithGithubEnv sets `github.env` from the given environment file. This is path of the temporary file that holds the
// environment variables
func (c *ExprContext) WithGithubEnv(path string) *ExprContext {
	c.Github.Env = path

	return c
}

// WithoutGithubEnv removes `github.env` from the context.
func (c *ExprContext) WithoutGithubEnv() *ExprContext {
	c.Github.Env = ""

	return c
}

// WithGithubPath sets `github.path` from the given environment file. This is path of the temporary file that holds the
func (c *ExprContext) WithGithubPath(path string) *ExprContext {
	c.Github.Path = path

	return c
}

// WithoutGithubPath removes `github.path` from the context.
func (c *ExprContext) WithoutGithubPath() *ExprContext {
	c.Github.Path = ""
	return c
}

// SetStepOutput sets the output of the given step.
func (c *ExprContext) SetStepOutput(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = gctx.StepContext{}
	}

	if sc.Outputs == nil {
		sc.Outputs = make(map[string]string)
	}

	sc.Outputs[key] = value

	c.Steps[stepID] = sc

	return c
}

// SetStepResult sets the result of the given step.
func (c *ExprContext) SetStepResult(stepID string, outcome, conclusion core.Conclusion) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = gctx.StepContext{}
	}

	sc.Outcome = outcome
	sc.Conclusion = conclusion

	c.Steps[stepID] = sc

	return c
}

// SetStepSummary sets the summary of the given step.
func (c *ExprContext) SetStepSummary(stepID, summary string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = gctx.StepContext{}
	}

	sc.Summary = summary

	c.Steps[stepID] = sc

	return c
}

// SetStepState sets the state of the given step.
func (c *ExprContext) SetStepState(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = gctx.StepContext{}
	}

	if sc.State == nil {
		sc.State = make(map[string]string)
	}

	sc.State[key] = value

	c.Steps[stepID] = sc

	return c
}

// SetJobStatus sets the status of the job.
func (c *ExprContext) SetJobStatus(status core.Conclusion) *ExprContext {
	c.Job.Status = status

	return c
}
