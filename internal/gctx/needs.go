package gctx

import "github.com/aweris/gale/internal/core"

// NeedContext is a context that contains information about dependent job.
type NeedContext struct {
	Result  core.Conclusion   `json:"result"`  // Result is conclusion of the job. It can be success, failure, skipped or cancelled.
	Outputs map[string]string `json:"outputs"` // Outputs of the job
}

type NeedsContext map[string]NeedContext

func (c *Context) LoadNeeds(runs ...core.JobRun) error {
	c.Needs = make(NeedsContext)

	for _, jr := range runs {
		c.Needs[jr.Job.ID] = NeedContext{Result: jr.Conclusion, Outputs: jr.Outputs}
	}

	return nil
}
