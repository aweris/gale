package preflight

import "github.com/aweris/gale/internal/core"

var _ Task = new(GHCheck)

type GHCheck struct{}

func (c *GHCheck) Name() string {
	return NameGHCheck
}

func (c *GHCheck) Type() TaskType {
	return TaskTypeCheck
}

func (c *GHCheck) DependsOn() []string {
	return []string{}
}

func (c *GHCheck) Run(_ *Context, _ Options) Result {
	// try to get token from github cli to make sure it is configured properly.
	_, err := core.GetToken()
	if err != nil {
		return Result{
			Status: Failed,
			Messages: []Message{
				{Level: Warning, Content: "GitHub CLI is exist or configured properly"},
				{Level: Error, Content: err.Error()},
			},
		}
	}

	return Result{Status: Passed, Messages: []Message{{Level: Info, Content: "GitHub CLI is exist and configured properly"}}}
}
