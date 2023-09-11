package gctx

import "github.com/aweris/gale/internal/core"

type RepoContext struct {
	*core.Repository
}

// LoadRepo initializes the context with the specified or current repository's default branch if no options are provided.
func (c *Context) LoadRepo(repo string, opts ...core.GetRepositoryOpts) error {
	r, err := core.GetRepository(repo, opts...)
	if err != nil {
		return err
	}

	c.Repo = RepoContext{r}

	return nil
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo(opts ...core.GetRepositoryOpts) error {
	r, err := core.GetCurrentRepository(opts...)
	if err != nil {
		return err
	}

	c.Repo = RepoContext{r}

	return nil
}
