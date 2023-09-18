package docker

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"dagger.io/dagger"
)

// Publish publishes the docker image for the given version.
func Publish(ctx context.Context, version string) error {
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	// ensure that the version is prefixed with a "v"
	version = strings.TrimPrefix(version, "v")
	version = fmt.Sprintf("v%s", version)

	image := fmt.Sprintf("ghcr.io/aweris/gale:%s", version)

	// If the registry is set, we'll use that instead of the default one. This is useful for testing and development.
	if registry := os.Getenv("_GALE_DOCKER_REGISTRY"); registry != "" {
		image = fmt.Sprintf("%s/aweris/gale:%s", registry, version)
	}

	var ldflags []string

	ldflags = append(ldflags, "-s", "-w")
	ldflags = append(ldflags, "-X github.com/aweris/gale/internal/version.gitVersion="+version)

	// builds

	gale := build(client, "./cmd/gale", "/src/out/gale")
	ghx := build(client, "./cmd/ghx", "/src/out/ghx")
	artifact := build(client, "./services/artifact", "/src/out/artifact-service")
	artifactCache := build(client, "./services/artifactcache", "/src/out/artifactcache-service")

	// container
	_, err = client.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "git", "docker", "github-cli"}).
		WithFile("/usr/local/bin/gale", gale).
		WithFile("/usr/local/bin/ghx", ghx).
		WithFile("/usr/local/bin/artifact-service", artifact).
		WithFile("/usr/local/bin/artifactcache-service", artifactCache).
		WithEntrypoint([]string{"/usr/local/bin/gale"}).
		Publish(ctx, image)

	return err
}

func build(client *dagger.Client, path, out string) *dagger.File {
	exec := []string{"go", "build", "-o", out, path}

	return client.Container().
		From("golang:"+strings.TrimPrefix(runtime.Version(), "go")).
		WithMountedDirectory("/src", client.Host().Directory(".")).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec(exec).
		File(out)
}
