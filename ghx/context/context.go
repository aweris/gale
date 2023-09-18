package context

import (
	"context"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/caarlos0/env/v9"

	"github.com/aweris/gale/internal/fs"
)

// Context represents the main context of the application.
type Context struct {
	Context   context.Context
	GhxConfig GhxConfig
	Dagger    DaggerContext
	Repo      RepositoryContext
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

	// load current repository information -- ghx executed from the repository root directory, using "." as path should be fine
	if err := ctx.Repo.LoadFromDirectory("."); err != nil {
		return nil, err
	}

	ctx.Github.Ref = ctx.Repo.Ref
	ctx.Github.RefName = ctx.Repo.RefName
	ctx.Github.RefType = string(ctx.Repo.RefType)
	ctx.Github.SHA = ctx.Repo.SHA

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
	ctx.Secrets.MountPath = filepath.Join(ctx.GhxConfig.HomeDir, "secrets", "secret.json")

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
