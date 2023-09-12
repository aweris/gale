package ghx

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/gctx"
)

// getStepName returns the step name. If step name is not set, it will be generated from the step type.
func getStepName(prefix string, s core.Step) string {
	if s.Name != "" {
		return strings.TrimSpace(strings.Join([]string{prefix, s.Name}, " "))
	}

	switch s.Type() {
	case core.StepTypeAction:
		return strings.TrimSpace(strings.Join([]string{prefix, s.Uses}, " "))
	case core.StepTypeRun:
		return strings.TrimSpace(strings.Join([]string{prefix, strings.Split(s.Run, "\n")[0]}, " "))
	default:
		return fmt.Sprintf("%s %s", prefix, s.ID)
	}
}

// evalStepCondition evaluates the given condition and returns the result. If the condition is empty, then it uses
// success() as default.
func evalStepCondition(condition string, ac *gctx.Context) (bool, core.Conclusion, error) {
	// if condition is empty, then use success() as default
	if condition == "" {
		condition = "success()"
	}

	// evaluate the condition as boolean expression
	run, err := expression.NewBoolExpr(condition).Eval(ac)
	if err != nil {
		return false, "", err
	}

	var conclusion core.Conclusion

	// if the condition is false, then set the conclusion as skipped
	if !run {
		conclusion = core.ConclusionSkipped
	}

	return run, conclusion, nil
}
