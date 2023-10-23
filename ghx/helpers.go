package main

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
	"github.com/aweris/gale/ghx/expression"
)

// getStepName returns the step name. If step name is not set, it will be generated from the step type.
func getStepName(prefix string, s model.Step) string {
	if s.Name != "" {
		return strings.TrimSpace(strings.Join([]string{prefix, s.Name}, " "))
	}

	switch s.Type() {
	case model.StepTypeAction:
		return strings.TrimSpace(strings.Join([]string{prefix, s.Uses}, " "))
	case model.StepTypeRun:
		return strings.TrimSpace(strings.Join([]string{prefix, strings.Split(s.Run, "\n")[0]}, " "))
	default:
		return fmt.Sprintf("%s %s", prefix, s.ID)
	}
}

// evalCondition evaluates the given condition and returns the result. If the condition is empty, then it uses
// success() as default.
func evalCondition(condition string, ac *context.Context) (bool, model.Conclusion, error) {
	// if condition is empty, then use success() as default
	if condition == "" {
		condition = "success()"
	}

	// evaluate the condition as boolean expression
	run, err := expression.NewBoolExpr(condition).Eval(ac)
	if err != nil {
		return false, "", err
	}

	var conclusion model.Conclusion

	// if the condition is false, then set the conclusion as skipped
	if !run {
		conclusion = model.ConclusionSkipped
	}

	return run, conclusion, nil
}
