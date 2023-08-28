package preflight

import (
	"fmt"
	"strings"
)

var _ Task = new(MissingExprContextCheck)

// MissingExprContextCheck is a preflight check that checks if there is any missing expression context.
type MissingExprContextCheck struct{}

func (m *MissingExprContextCheck) Name() string {
	return NameMissingExprContextCheck
}

func (m *MissingExprContextCheck) Type() TaskType {
	return TaskTypeCheck
}

func (m *MissingExprContextCheck) DependsOn() []string {
	return []string{NameDaggerCheck, NameWorkflowLoader}
}

func (m *MissingExprContextCheck) Run(ctx *Context, opt Options) Result {
	var (
		msg []Message

		status = Passed
	)

	data, err := ctx.Repo.GitRef.Dir.File(ctx.Workflow.Path).Contents(ctx.Context)
	if err != nil {
		return Result{
			Status: Failed,
			Messages: []Message{
				{Level: Error, Content: fmt.Sprintf("Load workflow file failed: %s", err.Error())},
			},
		}
	}

	if data == "" {
		return Result{
			Status: Failed,
			Messages: []Message{
				{Level: Error, Content: fmt.Sprintf("Workflow file is empty")},
			},
		}
	}

	// TODO: look better way other than hard-coding the list of not supported expression contexts.

	for _, context := range []string{"env", "vars", "strategy", "matrix", "needs", "jobs"} {
		if strings.Contains(data, fmt.Sprintf("${{ %s.", context)) {
			msg = append(msg, Message{Level: Warning, Content: fmt.Sprintf("Workflow file %s contains not supported expression context %s. This may cause unexpected behavior", ctx.Workflow.Path, context)})
		}
	}

	return Result{Status: status, Messages: msg}
}
