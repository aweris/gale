package gctx

// TODO: migrate to the new context slowly. These are added for keeping the backward compatibility while removing ExprContext.

// WithGithubEnv sets `github.env` from the given environment file. This is path of the temporary file that holds the
// environment variables
func (c *Context) WithGithubEnv(path string) *Context {
	c.Github.Env = path

	return c
}

// WithoutGithubEnv removes `github.env` from the context.
func (c *Context) WithoutGithubEnv() *Context {
	c.Github.Env = ""

	return c
}

// WithGithubPath sets `github.path` from the given environment file. This is path of the temporary file that holds the
func (c *Context) WithGithubPath(path string) *Context {
	c.Github.Path = path

	return c
}

// WithoutGithubPath removes `github.path` from the context.
func (c *Context) WithoutGithubPath() *Context {
	c.Github.Path = ""
	return c
}
