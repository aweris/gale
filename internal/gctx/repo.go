package gctx

import (
	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/data"
)

type LoadRepoOpts struct {
	Branch       string
	Tag          string
	WorkflowsDir string
}

type RepoContext struct {
	Repository   *core.Repository  `json:"-"`
	CacheVol     *data.CacheVolume `json:"-"`
	WorkflowsDir string            `json:"workflows_dir" env:"GALE_WORKFLOWS_DIR" envDefault:".github/workflows" container_env:"true"`
}

// LoadRepo initializes the context with the specified or current repository's default branch if no options are provided.
func (c *Context) LoadRepo(repo string, opts ...LoadRepoOpts) error {
	rc, err := NewContextFromEnv[RepoContext]()
	if err != nil {
		return err
	}

	opt := LoadRepoOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	r, err := core.GetRepository(repo, core.GetRepositoryOpts{Branch: opt.Branch, Tag: opt.Tag})
	if err != nil {
		return err
	}

	rc.Repository = r

	// cache volume only used non container mode. Instead of leaving it nil, we create a new one in container mode kas
	// well to avoid nil pointer errors. There is no harm in doing so since we're supporting Dagger in Dagger mode.
	rc.CacheVol = data.NewCacheVolume(r)

	// if it's not in container mode, workflows dir should be set from the options
	if !c.isContainer {
		rc.WorkflowsDir = opt.WorkflowsDir
	}

	c.Repo = rc

	return nil
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo(opts ...LoadRepoOpts) error {
	return c.LoadRepo("", opts...)
}

var _ helpers.WithContainerFuncHook = new(RepoContext)

func (c RepoContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return WithContainerEnv[RepoContext](config.Client(), c)(container)
	}
}
