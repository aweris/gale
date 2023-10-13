package gctx

import (
	"context"
	"fmt"
	"math"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/expression"
)

type Context struct {
	debug     bool             // debug indicates whether the workflow is running in debug mode.
	path      string           // path is the data path for the context to be mounted from the host or to be used in the container.
	Context   context.Context  // Context is the current context of the workflow.
	Client    *dagger.Client   // Client is the dagger client to be used in the workflow.
	Repo      RepoContext      // Repo is the context for the repository.
	Execution ExecutionContext // Execution is the context for the execution.
	Actions   ActionsContext   // Actions is the context for the actions.

	// Github Expression Contexts
	Runner  RunnerContext
	Github  GithubContext
	Secrets SecretsContext
	Inputs  InputsContext
	Job     JobContext
	Steps   StepsContext
	Needs   NeedsContext
	Matrix  core.MatrixCombination
	Env     map[string]string
}

func Load(ctx context.Context, debug bool, client *dagger.Client) (*Context, error) {
	gctx := &Context{debug: debug, Context: ctx, Client: client, path: "/home/runner/work/_temp/gale"}

	// load actions context
	err := gctx.LoadActionsContext()
	if err != nil {
		return nil, err
	}

	// load the repository context
	err = gctx.LoadRunnerContext()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadGithubContext()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadSecrets()
	if err != nil {
		return nil, err
	}

	err = gctx.LoadInputs()
	if err != nil {
		return nil, err
	}

	gctx.Secrets.setToken(gctx.Github.Token)

	return gctx, nil
}

// expression.VariableProvider interface to be used in expressions.

var _ expression.VariableProvider = new(Context)

func (c *Context) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return c.Github, nil
	case "runner":
		return c.Runner, nil
	case "env":
		return c.Env, nil
	case "vars":
		return map[string]string{}, nil
	case "job":
		return c.Job, nil
	case "steps":
		return c.Steps, nil
	case "secrets":
		return c.Secrets.Data, nil
	case "strategy":
		return map[string]string{}, nil
	case "matrix":
		return c.Matrix, nil
	case "needs":
		return c.Needs, nil
	case "inputs":
		return c.Inputs, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	default:
		return nil, fmt.Errorf("unknown variable: %s", name)
	}
}
