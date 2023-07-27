package actions

import (
	"fmt"
	"math"

	"github.com/aweris/gale/tools/ghx/expression"
)

var _ expression.VariableProvider = new(ExprContext)

type ExprContext struct {
	// TODO: add other contexts when needed.
	// - github context
	// - runner context
	// - env context
	// - job context
	// - steps context
	//  - vars context
	//  - secrets context
	//  - strategy context
	//  - matrix context
	//  - needs context
	//  - jobs context
	//  - inputs context
}

func (c *ExprContext) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return map[string]string{}, nil
	case "runner":
		return map[string]string{}, nil
	case "env":
		return map[string]string{}, nil
	case "vars":
		return map[string]string{}, nil
	case "job":
		return map[string]string{}, nil
	case "steps":
		return map[string]string{}, nil
	case "secrets":
		return map[string]string{}, nil
	case "strategy":
		return map[string]string{}, nil
	case "matrix":
		return map[string]string{}, nil
	case "needs":
		return map[string]string{}, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	default:
		return nil, fmt.Errorf("unknown variable: %s", name)
	}
}
