package services

import (
	"context"
	"fmt"

	"os"

	"dagger.io/dagger"

	"golang.org/x/mod/semver"

	"github.com/magefile/mage/mg"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/images"
)

type Artifact mg.Namespace

// Publish publishes the artifact service with the given version.
func (_ Artifact) Publish(ctx context.Context, version string) error {
	if version != "main" {
		if ok := semver.IsValid(version); !ok {
			return fmt.Errorf("invalid semver tag: %s", version)
		}
	}

	image := fmt.Sprintf("ghcr.io/aweris/gale/services/artifact:%s", version)
	// If the registry is set, we'll use that instead of the default one. This is useful for testing and development.
	if registry := os.Getenv("_GALE_DOCKER_REGISTRY"); registry != "" {
		image = fmt.Sprintf("%s/aweris/gale/services/artifact:%s", registry, version)
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	// use config.Client() instead of client just to keep the code consistent with in repo
	config.SetClient(client)

	file := images.GoBase().
		WithMountedDirectory("/src", client.Host().Directory(".")).
		WithWorkdir("/src/services/artifact").
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
