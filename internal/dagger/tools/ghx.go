package tools

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/version"
)

// Ghx returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func Ghx(ctx context.Context) (*dagger.File, error) {
	v := version.GetVersion()

	tag := v.GitVersion

	file := config.Client().Container().From("ghcr.io/aweris/gale/tools/ghx:" + tag).File("/ghx")

	// check, if the file doesn't exist or is empty
	if size, err := file.Size(ctx); size == 0 || err != nil {
		return nil, fmt.Errorf("ghx@%s binary not available", tag)
	}

	return file, nil
}
