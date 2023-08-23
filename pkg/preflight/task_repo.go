package preflight

import (
	"fmt"

	"github.com/aweris/gale/internal/core"
)

var _ Task = new(RepoLoader)

type RepoLoader struct{}

func (c *RepoLoader) Name() string {
	return NameRepoLoader
}

func (c *RepoLoader) Type() TaskType {
	return TaskTypeLoad
}

func (c *RepoLoader) DependsOn() []string {
	return []string{NameDaggerCheck, NameGHCheck}
}

func (c *RepoLoader) Run(ctx *Context, opts Options) Result {
	var msg []Message

	if opts.Repo != "" {
		repo, err := core.GetRepository(opts.Repo, core.GetRepositoryOpts{Branch: opts.Branch, Tag: opts.Tag})
		if err != nil {
			return Result{
				Status: Failed,
				Messages: []Message{
					{Level: Error, Content: fmt.Sprintf("Get repository %s failed: %s", opts.Repo, err.Error())},
				},
			}
		}

		ctx.Repo = repo

		msg = append(msg, Message{Level: Info, Content: fmt.Sprintf("Repository %s is loaded", opts.Repo)})
	} else {
		repo, err := core.GetCurrentRepository()
		if err != nil {
			return Result{
				Status: Failed,
				Messages: []Message{
					{Level: Error, Content: fmt.Sprintf("Get current repository failed: %s", err.Error())},
				},
			}
		}

		ctx.Repo = repo

		msg = append(msg, Message{Level: Info, Content: fmt.Sprintf("Current repository %s is loaded", repo.Name)})
	}

	return Result{Status: Passed, Messages: msg}
}
