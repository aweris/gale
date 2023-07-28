package actions

import (
	"fmt"
	"math"
	"os"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/tools/ghx/expression"
)

var _ expression.VariableProvider = new(ExprContext)

type ExprContext struct {
	Github GithubContext // Github context

	// TODO: add other contexts when needed.
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

func NewExprContext() *ExprContext {
	return &ExprContext{
		Github: GithubContext{
			GithubRepositoryContext: core.GithubRepositoryContext{
				Repository:        os.Getenv("GITHUB_REPOSITORY"),
				RepositoryID:      os.Getenv("GITHUB_REPOSITORY_ID"),
				RepositoryOwner:   os.Getenv("GITHUB_REPOSITORY_OWNER"),
				RepositoryOwnerID: os.Getenv("GITHUB_REPOSITORY_OWNER_ID"),
				RepositoryURL:     os.Getenv("GITHUB_REPOSITORY_URL"),
				Workspace:         os.Getenv("GITHUB_WORKSPACE"),
			},
			GithubSecretsContext: core.GithubSecretsContext{
				Token: os.Getenv("GITHUB_TOKEN"),
			},
		},
	}
}

// GithubContext contains information about the workflow run and the event that triggered the run and event that
// triggered the run.
//
// Contents of this context are managed by sub-contexts. This is just a composite context to provide variables for
// expressions.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type GithubContext struct {
	core.GithubRepositoryContext
	core.GithubSecretsContext

	// TODO: add missing contexts when needed.
}

func (c *ExprContext) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return c.Github, nil
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
