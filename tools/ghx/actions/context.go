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
	Github GithubContext               // Github context
	Steps  map[string]core.StepContext // Steps context

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
			GithubURLContext: core.GithubURLContext{
				ApiURL:     os.Getenv("GITHUB_API_URL"),
				GraphqlURL: os.Getenv("GITHUB_GRAPHQL_URL"),
				ServerURL:  os.Getenv("GITHUB_SERVER_URL"),
			},
			GithubFilesContext: core.GithubFilesContext{ /* No initial values */ },
		},
		Steps: make(map[string]core.StepContext),
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
	// Global contexts - these are applied to container and available to all steps.
	core.GithubRepositoryContext
	core.GithubSecretsContext
	core.GithubURLContext

	// Local contexts - these contexts changes at course of the workflow run.
	core.GithubFilesContext

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
		return c.Steps, nil
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

// WithGithubEnv sets `github.env` from the given environment file. This is path of the temporary file that holds the
// environment variables
func (c *ExprContext) WithGithubEnv(ef *core.EnvironmentFile) *ExprContext {
	c.Github.GithubFilesContext.Env = ef.Path

	return c
}

// WithoutGithubEnv removes `github.env` from the context.
func (c *ExprContext) WithoutGithubEnv() *ExprContext {
	c.Github.GithubFilesContext.Env = ""

	return c
}

// WithGithubPath sets `github.path` from the given environment file. This is path of the temporary file that holds the
func (c *ExprContext) WithGithubPath(ef *core.EnvironmentFile) *ExprContext {
	c.Github.GithubFilesContext.Path = ef.Path

	return c
}

// WithoutGithubPath removes `github.path` from the context.
func (c *ExprContext) WithoutGithubPath() *ExprContext {
	c.Github.GithubFilesContext.Path = ""
	return c
}

// SetStepOutput sets the output of the given step.
func (c *ExprContext) SetStepOutput(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	if sc.Outputs == nil {
		sc.Outputs = make(map[string]string)
	}

	sc.Outputs[key] = value

	c.Steps[stepID] = sc

	return c
}

// SetStepResult sets the result of the given step.
func (c *ExprContext) SetStepResult(stepID string, outcome, conclusion core.Conclusion) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	sc.Outcome = outcome
	sc.Conclusion = conclusion

	c.Steps[stepID] = sc

	return c
}

// SetStepSummary sets the summary of the given step.
func (c *ExprContext) SetStepSummary(stepID, summary string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	sc.Summary = summary

	c.Steps[stepID] = sc

	return c
}

// SetStepState sets the state of the given step.
func (c *ExprContext) SetStepState(stepID, key, value string) *ExprContext {
	sc, ok := c.Steps[stepID]
	if !ok {
		sc = core.StepContext{}
	}

	if sc.State == nil {
		sc.State = make(map[string]string)
	}

	sc.State[key] = value

	c.Steps[stepID] = sc

	return c
}
