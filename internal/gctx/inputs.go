package gctx

// InputsContext contains input properties passed to an action, to a reusable workflow, or to a manually triggered
// workflow.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context
type InputsContext map[string]string

func (c *Context) LoadInputs() error {
	c.Inputs = make(InputsContext)

	return nil
}
