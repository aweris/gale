package gctx

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/pkg/data"
)

// SecretsContext is a context that contains secrets.
type SecretsContext struct {
	MountPath string            // MountPath is the path where the secrets are mounted.
	Data      map[string]string // Data is the secrets data.
}

// LoadSecrets loads the secrets into the context. If the context is in container mode, it will read the secrets from
// the mounted file. Otherwise, it will load the secrets from the given maps.
func (c *Context) LoadSecrets(secrets ...map[string]string) error {
	c.Secret = SecretsContext{MountPath: filepath.Join(c.path, data.DirSecrets, "secret.json")}

	if !c.isContainer {
		sd := make(map[string]string)

		// merge all secrets into a single optional map.
		for _, s := range secrets {
			for k, v := range s {
				sd[k] = v
			}
		}

		c.Secret.Data = sd

		return nil
	}

	// if it's in container mode, we need to read the secrets from the mounted file.

	err := fs.EnsureFile(c.Secret.MountPath)
	if err != nil {
		return fmt.Errorf("failed to ensure secrets file exist: %w", err)
	}

	err = fs.ReadJSONFile(c.Secret.MountPath, &c.Secret.Data)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	return nil
}

// helpers.WithContainerFuncHook interface to be loaded in the container.

var _ helpers.WithContainerFuncHook = new(SecretsContext)

func (c *SecretsContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		raw, err := json.Marshal(c.Data)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		secret := config.Client().SetSecret("secrets-context", string(raw))

		return container.WithMountedSecret(c.MountPath, secret)
	}
}

// SetToken sets the token in the context.
func (c *SecretsContext) SetToken(token string) *SecretsContext {
	c.Data["GITHUB_TOKEN"] = token // GITHUB_TOKEN is a special secret that is always available to the workflow.

	return c
}

// SetSecrets sets the secrets to the context.
func (c *SecretsContext) SetSecrets(secrets map[string]string) *SecretsContext {
	for k, v := range secrets {
		c.Data[k] = v
	}

	return c
}
