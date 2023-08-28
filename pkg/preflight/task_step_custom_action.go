package preflight

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
)

var _ Task = new(StepActionCheck)

// StepActionCheck is a preflight check that checks related to step action.
type StepActionCheck struct{}

func (s *StepActionCheck) Name() string {
	return NameStepActionCheck
}

func (s *StepActionCheck) Type() TaskType {
	return TaskTypeCheck
}

func (s *StepActionCheck) DependsOn() []string {
	return []string{NameDaggerCheck, NameWorkflowLoader}
}

func (s *StepActionCheck) Run(ctx *Context, _ Options) Result {
	var (
		status    = Passed
		checkNode = false
		msg       = make([]Message, 0)
	)

	// runner container
	runner := config.Client().Container().From(config.RunnerImage())

	for uses, action := range ctx.CustomActions {
		if action.Meta.Runs.Using == core.ActionRunsUsingComposite {
			status = Failed
			msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Action %s is is a composite action. Composite actions are not supported", uses)})
			continue
		}

		if action.Meta.Runs.Using == core.ActionRunsUsingNode12 || action.Meta.Runs.Using == core.ActionRunsUsingNode16 {
			checkNode = true
		}
	}

	if checkNode {
		out, err := runner.WithExec([]string{"which", "node"}).Stdout(ctx.Context)

		// ignore the error since we will check the output, not the error. This check just to ensure we'll not hide
		// any error.
		if err != nil {
			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Check node returned error: %s", err.Error())})
		}

		// trim the output and remove the new line character
		out = strings.TrimSpace(strings.TrimSuffix(out, "\n"))

		// if the output is empty or contains not found message then the shell is not available
		if out == "" || out == "node not found" {
			status = Failed
			msg = append(msg, Message{Level: Error, Content: "Node is not available in the runner image"})
		} else {
			msg = append(msg, Message{Level: Debug, Content: "Node is available in the runner image"})

			version, err := runner.WithExec([]string{"node", "--version"}).Stdout(ctx.Context)

			// trim the output and remove the new line character
			version = strings.TrimSpace(strings.TrimSuffix(version, "\n"))

			if err != nil {
				msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Check node version returned error: %s", err.Error())})
			} else {
				msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Node version is %s", version)})
			}
		}
	}

	return Result{Status: status, Messages: msg}
}
