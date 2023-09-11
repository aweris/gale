package gctx

import (
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/pkg/data"
)

type RepoContext struct {
	Repository *core.Repository
	CacheVol   *data.CacheVolume
}

// LoadRepo initializes the context with the specified or current repository's default branch if no options are provided.
func (c *Context) LoadRepo(repo string, opts ...core.GetRepositoryOpts) error {
	r, err := core.GetRepository(repo, opts...)
	if err != nil {
		return err
	}

	c.Repo = RepoContext{Repository: r}

	// cache volume only used non container mode. Instead of leaving it nil, we create a new one in container mode as
	// well to avoid nil pointer errors. There is no harm in doing so since we're supporting Dagger in Dagger mode.
	c.Repo.CacheVol = data.NewCacheVolume(r)

	return nil
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo(opts ...core.GetRepositoryOpts) error {
	return c.LoadRepo("", opts...)
}
