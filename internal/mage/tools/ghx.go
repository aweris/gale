package tools

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"golang.org/x/mod/semver"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/dagger/images"
)

type Ghx mg.Namespace

// Publish publishes the artifact service to ghcr.io/aweris/gale/services/artifact with the given version.
func (_ Ghx) Publish(ctx context.Context, version string) error {
	if version != "main" {
		if ok := semver.IsValid(version); !ok {
			return fmt.Errorf("invalid semver tag: %s", version)
		}
	}

	image := fmt.Sprintf("ghcr.io/aweris/gale/tools/ghx:%s", version)

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}

	file := images.GoBase(client).
		WithMountedDirectory("/src", client.Host().Directory(".")).
		WithWorkdir("/src/tools/ghx").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "/src/out/ghx", "."}).
		File("/src/out/ghx")

	addr, err := client.Container().
		From("gcr.io/distroless/static").
		WithFile("/ghx", file).
		WithEntrypoint([]string{"/ghx"}).
		Publish(ctx, image)
	if err != nil {
		return err
	}

	fmt.Printf("Artifact service published to %s\n", addr)

	return nil
}
