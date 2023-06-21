package binaries

import (
	"context"
	"fmt"

	"dagger.io/dagger"
)

// Ghx returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func Ghx(ctx context.Context, client *dagger.Client, version string) (*dagger.File, error) {
	file := client.Container().
		Pipeline("Binaries").Pipeline("ghx").
		From(fmt.Sprintf("ghcr.io/aweris/ghx:%s", version)).
		Directory("/usr/local/bin").
		File("ghx")

	// check, if the file doesn't exist or is empty
	if size, err := file.Size(ctx); size == 0 || err != nil {
		return nil, fmt.Errorf("ghx@%s binary not available", version)
	}

	return file, nil
}
