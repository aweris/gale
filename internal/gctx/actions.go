package gctx

// ActionsContext is the context for the internal services configuration for used by GitHub Actions.
type ActionsContext struct {
	RuntimeURL string `env:"ACTIONS_RUNTIME_URL"`
	CacheURL   string `env:"ACTIONS_CACHE_URL"`
	Token      string `env:"ACTIONS_RUNTIME_TOKEN"`
}

func (c *Context) LoadActionsContext() error {
	ac, err := NewContextFromEnv[ActionsContext]()
	if err != nil {
		return err
	}

	c.Actions = ac

	return nil
}
