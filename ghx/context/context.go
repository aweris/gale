package context

import (
	"context"

	"dagger.io/dagger"

	"github.com/caarlos0/env/v9"

	"github.com/aweris/gale/common/fs"
)

// Context represents the main context of the application.
type Context struct {
	Context   context.Context
	GhxConfig GhxConfig
	Dagger    DaggerContext
	Execution ExecutionContext
	Actions   ActionsContext
	Github    GithubContext
	Inputs    InputsContext
	Job       JobContext
	Needs     NeedsContext
	Runner    RunnerContext
	Secrets   SecretsContext
	Steps     StepsContext
	Env       EnvContext
	Matrix    MatrixContext
}

// New returns a new Context initialized from environment variables.
func New(std context.Context, client *dagger.Client) (*Context, error) {
	var ctx Context

	if err := env.Parse(&ctx); err != nil {
		return nil, err
	}

	// init empty custom type contexts
	if ctx.Inputs == nil {
		ctx.Inputs = make(InputsContext)
	}

	if ctx.Needs == nil {
		ctx.Needs = make(NeedsContext)
	}

	if ctx.Steps == nil {
		ctx.Steps = make(StepsContext)
	}

	if ctx.Env == nil {
		ctx.Env = make(EnvContext)
	}

	if ctx.Matrix == nil {
		ctx.Matrix = make(MatrixContext)
	}

	// set the standard ctx
	ctx.Context = std

	// set the dagger client
	ctx.Dagger.Client = client

	// set non environment config for github ctx
	if ctx.Github.EventPath != "" {
		err := fs.ReadJSONFile(ctx.Github.EventPath, &ctx.Github.Event)
		if err != nil {
			return nil, err
		}
	}

	// just to be sure that we have a initialized event map
	if ctx.Github.Event == nil {
		ctx.Github.Event = make(map[string]interface{})
	}

	// set secrets ctx
	secretsMountPath, err := ctx.GetSecretsPath()

	ctx.Secrets.MountPath = secretsMountPath
	if err != nil {
		return nil, err
	}

	if err := fs.EnsureFile(ctx.Secrets.MountPath); err != nil {
		return nil, err
	}

	if err := fs.ReadJSONFile(ctx.Secrets.MountPath, &ctx.Secrets.Data); err != nil {
		return nil, err
	}

	// ensure secrets data is initialized
	if ctx.Secrets.Data == nil {
		ctx.Secrets.Data = make(map[string]string)
	}

	// add github token to secrets
	ctx.Secrets.Data["GITHUB_TOKEN"] = ctx.Github.Token

	// update environment variables with defaults and manually set values
	syncWithEnvValues(&ctx)

	return &ctx, nil
}

// Debug returns true if debug mode is enabled.
func (c *Context) Debug() bool {
	return c.Runner.Debug == "1"
}
