package gctx

import "github.com/aweris/gale/internal/core"

// JobContext contains information about the currently running job.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
type JobContext struct {
	Status core.Conclusion `json:"status"` // Status is the current status of the job. Possible values are success, failure, or cancelled.

	// TODO: add other fields when needed.
}

func (c *Context) LoadJob(status core.Conclusion) error {
	c.Job = JobContext{
		Status: status,
	}

	return nil
}
