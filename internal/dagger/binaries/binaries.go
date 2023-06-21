package binaries

import (
	"context"
	"fmt"

	"dagger.io/dagger"

	"github.com/Masterminds/semver/v3"

	"github.com/aweris/gale/internal/dagger/images"
)

// Ghx returns a dagger file for the ghx binary. It'll return an error if the binary is not available.
func Ghx(ctx context.Context, client *dagger.Client, version string) (*dagger.File, error) {
	var file *dagger.File

	// if the version is not a valid semver, we'll assume that it's a branch name or commit hash,
	// and we'll try to install using go
	_, err := semver.NewVersion(version)
	if err != nil {
		file = images.GoBase(client).
			Pipeline("Install ghx binary").
			WithExec([]string{"go", "install", fmt.Sprintf("github.com/aweris/ghx@%s", version)}).
			Directory("/go/bin").
			File("ghx")
	} else {
		file = client.Container().
			Pipeline("Download ghx binary").
			From(fmt.Sprintf("ghcr.io/aweris/ghx:%s", version)).
			Directory("/usr/local/bin").
			File("ghx")
	}

	// check, if the file doesn't exist or is empty
	if size, err := file.Size(ctx); size == 0 || err != nil {
		return nil, fmt.Errorf("ghx@%s binary not available", version)
	}

	return file, nil
}
