package gctx

import (
	"dagger.io/dagger"
	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

// ActionsContext is the context for the internal services configuration for used by GitHub Actions.
type ActionsContext struct {
	RuntimeURL string `env:"ACTIONS_RUNTIME_URL" container_env:"true"`
	CacheURL   string `env:"ACTIONS_CACHE_URL" container_env:"true"`
	Token      string `env:"ACTIONS_RUNTIME_TOKEN" container_env:"true"`
}

func (c *Context) LoadActionsContext() error {
	ac, err := NewContextFromEnv[ActionsContext]()
	if err != nil {
		return err
	}

	c.Actions = ac

	return nil
}

// helpers.WithContainerFuncHook interface to be loaded in the container.

var _ helpers.WithContainerFuncHook = new(ActionsContext)

func (c *ActionsContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// Load context in the container as environment variables or secrets.
		container = container.With(WithContainerEnv(config.Client(), c))

		return container
	}
}
