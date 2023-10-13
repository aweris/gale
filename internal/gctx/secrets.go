package gctx

import (
	"fmt"
	"path/filepath"

	"github.com/aweris/gale/internal/fs"
)

// SecretsContext is a context that contains secrets.
type SecretsContext struct {
	MountPath string            // MountPath is the path where the secrets are mounted.
	Data      map[string]string // Data is the secrets data.
}

// LoadSecrets loads the secrets into the context. If the context is in container mode, it will read the secrets from
// the mounted file. Otherwise, it will load the secrets from the given maps.
func (c *Context) LoadSecrets() error {
	c.Secrets = SecretsContext{MountPath: filepath.Join(c.path, "secrets", "secret.json")}

	// if it's in container mode, we need to read the secrets from the mounted file.
	err := fs.EnsureFile(c.Secrets.MountPath)
	if err != nil {
		return fmt.Errorf("failed to ensure secrets file exist: %w", err)
	}

	err = fs.ReadJSONFile(c.Secrets.MountPath, &c.Secrets.Data)
	if err != nil {
		return fmt.Errorf("failed to read secrets file: %w", err)
	}

	return nil
}

// setToken sets the token in the context.
func (c *SecretsContext) setToken(token string) {
	c.Data["GITHUB_TOKEN"] = token // GITHUB_TOKEN is a special secret that is always available to the workflow.
}
