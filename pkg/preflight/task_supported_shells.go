package preflight

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/internal/config"
)

var _ Task = new(StepShellCheck)

// StepShellCheck is a preflight check that checks if step shell is supported.
type StepShellCheck struct{}

func (s *StepShellCheck) Name() string {
	return NameStepShellCheck
}

func (s *StepShellCheck) Type() TaskType {
	return TaskTypeCheck
}

func (s *StepShellCheck) DependsOn() []string {
	return []string{NameDockerImagesCheck, NameWorkflowLoader}
}

func (s *StepShellCheck) Run(ctx *Context, _ Options) Result {
	var (
		msg []Message

		status = Passed
	)

	// TODO: look better way other than hard-coding the list of not supported shells.

	// list of not supported shells. This list will be updated when a new shell is supported.
	notSupportedShells := map[string]bool{
		"pwsh":       true,
		"powershell": true,
		"cmd":        true,
	}

	// runner container
	runner := config.Client().Container().From(config.RunnerImage())

	for shell, _ := range ctx.Shells {
		if _, ok := notSupportedShells[shell]; ok {
			status = Failed
			msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Shell %s is not supported", shell)})
			continue
		}

		// TODO: not sure if this is the best way to check if a shell is available in the runner image.
		//  however, this will work for current runner image.

		// check if we can execute a container with given dagger context exist in environment
		out, err := runner.WithExec([]string{"which", shell}).Stdout(ctx.Context)
		// ignore the error since we will check the output, not the error. This check just to ensure we'll not hide
		// any error.
		if err != nil {
			msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Check shell %s returned error: %s", shell, err.Error())})
		}

		// trim the output and remove the new line character
		out = strings.TrimSpace(strings.TrimSuffix(out, "\n"))

		// if the output is empty or contains not found message then the shell is not available
		if out == "" || out == fmt.Sprintf("%s not found", shell) {
			status = Failed
			msg = append(msg, Message{Level: Error, Content: fmt.Sprintf("Shell %s is not available in the runner image", shell)})
			continue
		}

		msg = append(msg, Message{Level: Debug, Content: fmt.Sprintf("Shell %s is available in the runner image", out)})
	}

	return Result{Status: status, Messages: msg}
}
