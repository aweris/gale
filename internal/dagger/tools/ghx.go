package tools

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/version"
)

// Ghx returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func Ghx(ctx context.Context, client *dagger.Client) (*dagger.File, error) {
	v := version.GetVersion()

	tag := v.GitVersion

	// If the tag is a dev tag, we'll use the main branch.
	if tag == "v0.0.0-dev" {
		tag = "main"
	}

	file := client.Container().From("ghcr.io/aweris/gale/tools/ghx:" + tag).File("/ghx")

	// check, if the file doesn't exist or is empty
	if size, err := file.Size(ctx); size == 0 || err != nil {
		return nil, fmt.Errorf("ghx@%s binary not available", tag)
	}

	return file, nil
}
