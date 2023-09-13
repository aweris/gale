package gctx

import "github.com/aweris/gale/internal/core"

// TODO: migrate to the new context slowly. These are added for keeping the backward compatibility while removing ExprContext.

// WithGithubEnv sets `github.env` from the given environment file. This is path of the temporary file that holds the
// environment variables
func (c *Context) WithGithubEnv(path string) *Context {
	c.Github.Env = path

	return c
}

// WithoutGithubEnv removes `github.env` from the context.
func (c *Context) WithoutGithubEnv() *Context {
	c.Github.Env = ""

	return c
}

// WithGithubPath sets `github.path` from the given environment file. This is path of the temporary file that holds the
func (c *Context) WithGithubPath(path string) *Context {
	c.Github.Path = path

	return c
}

// WithoutGithubPath removes `github.path` from the context.
func (c *Context) WithoutGithubPath() *Context {
	c.Github.Path = ""
	return c
}

// SetStepResult sets the result of the given step.
func (c *Context) SetStepResult(stepID string, outcome, conclusion core.Conclusion) *Context {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = StepContext{}
	}

	sc.Outcome = outcome
	sc.Conclusion = conclusion

	c.Steps[stepID] = sc

	return c
}

// SetJobStatus sets the status of the job.
func (c *Context) SetJobStatus(status core.Conclusion) *Context {
	c.Job.Status = status

	return c
}
