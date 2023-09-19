package docker

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"dagger.io/dagger"

	"github.com/google/uuid"
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

	// build ldflags for commands
	var ldflags []string

	ldflags = append(ldflags, "-s", "-w")
	ldflags = append(ldflags, "-X github.com/aweris/gale/internal/version.gitVersion="+version)

	// builds all components of the gale

	gale := build(client, "./cmd/gale", ldflags...)
	ghx := build(client, "./cmd/ghx", ldflags...)
	artifact := build(client, "./services/artifact")
	artifactCache := build(client, "./services/artifactcache")

	// create a container that will be used to publish the image
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

// build builds the code for the given path and returns the output file.
func build(client *dagger.Client, path string, ldflags ...string) *dagger.File {
	out := uuid.New().String()

	exec := []string{"go", "build", "-o", out}

	if len(ldflags) > 0 {
		exec = append(exec, "-ldflags")
		exec = append(exec, strings.Join(ldflags, " "))
	}

	exec = append(exec, path)

	return client.Container().
		From("golang:"+strings.TrimPrefix(runtime.Version(), "go")).
		WithMountedDirectory("/src", client.Host().Directory(".")).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0").
		WithExec(exec).
		File(out)
}
