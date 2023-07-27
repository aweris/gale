package runner

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/internal/core"
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
