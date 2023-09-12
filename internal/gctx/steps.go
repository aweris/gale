package gctx

import "github.com/aweris/gale/internal/core"

// StepContext is a context that contains information about the step.
//
// This context created per step execution. It won't be used or applied to the container level.
type StepContext struct {
	Conclusion core.Conclusion   `json:"conclusion"` // Conclusion is the result of a completed step after continue-on-error is applied
	Outcome    core.Conclusion   `json:"outcome"`    // Outcome is  the result of a completed step before continue-on-error is applied
	Outputs    map[string]string `json:"outputs"`    // Outputs is a map of output name to output value
	State      map[string]string `json:"-"`          // State is a map of step state variables. This is not available to expressions so that's why json tag is set to "-" to ignore it.
	Summary    string            `json:"-"`          // Summary is the summary of the step. This is not available to expressions so that's why json tag is set to "-" to ignore it.
}

type StepsContext map[string]StepContext

func (c *Context) LoadSteps() error {
	c.Steps = make(StepsContext)

	return nil
}
