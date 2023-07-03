package services

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"golang.org/x/mod/semver"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/dagger/images"
)

type Artifact mg.Namespace

// Publish publishes the artifact service to ghcr.io/aweris/gale/services/artifact with the given version.
func (_ Artifact) Publish(ctx context.Context, version string) error {
	if version != "main" {
		if ok := semver.IsValid(version); !ok {
			return fmt.Errorf("invalid semver tag: %s", version)
		}
	}

	image := fmt.Sprintf("ghcr.io/aweris/gale/services/artifact:%s", version)

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}

	file := images.GoBase(client).
		WithMountedDirectory("/src", client.Host().Directory("services/artifact")).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec([]string{"go", "build", "-o", "/src/out/artifact-service", "."}).
		File("/src/out/artifact-service")

	addr, err := client.Container().
		From("gcr.io/distroless/static").
		WithFile("/entrypoint", file).
		WithEntrypoint([]string{"/entrypoint"}).
		Publish(ctx, image)
	if err != nil {
		return err
	}

	fmt.Printf("Artifact service published to %s\n", addr)

	return nil
}
