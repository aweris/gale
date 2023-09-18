package context

import (
	"fmt"
	"math"

	"github.com/aweris/gale/ghx/expression"
)

// expression.VariableProvider interface to be used in expressions.
var _ expression.VariableProvider = new(Context)

func (c *Context) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return c.Github, nil
	case "runner":
		return c.Runner, nil
	case "env":
		return c.Env, nil
	case "vars":
		return map[string]string{}, nil
	case "job":
		return c.Job, nil
	case "steps":
		return c.Steps, nil
	case "secrets":
		return c.Secrets.Data, nil
	case "strategy":
		return map[string]string{}, nil
	case "matrix":
		return c.Matrix, nil
	case "needs":
		return c.Needs, nil
	case "inputs":
		return c.Inputs, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	default:
		return nil, fmt.Errorf("unknown variable: %s", name)
	}
}

var _ expression.VariableProvider = new(ActionsVariableProvider)

// ActionsVariableProvider is a variable provider for actions. It provides the inputs context contains the inputs of the
// current action to pass to the expressions. All other variables are provided by the main context.
type ActionsVariableProvider struct {
	main   *Context
	inputs InputsContext
}

// GetVariableProvider returns a variable provider for the current action. If the current action or step run is nil,
// it returns the main context as the variable provider.
func (c *Context) GetVariableProvider() expression.VariableProvider {
	if c.Execution.StepRun == nil || c.Execution.CurrentAction == nil {
		return c
	}

	var (
		inputs = make(InputsContext)
		step   = c.Execution.StepRun.Step
		action = c.Execution.CurrentAction
	)

	for k, v := range step.With {
		inputs[k] = v
	}

	// add default values for inputs that are not defined in the step config
	for k, v := range action.Meta.Inputs {
		if _, ok := step.With[k]; ok {
			continue
		}

		if v.Default == "" {
			continue
		}

		inputs[k] = v.Default
	}

	return &ActionsVariableProvider{main: c, inputs: inputs}
}

func (p *ActionsVariableProvider) GetVariable(name string) (interface{}, error) {
	if name == "inputs" {
		return p.inputs, nil
	}

	return p.main.GetVariable(name)
}
