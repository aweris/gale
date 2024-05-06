package main

import (
	"context"
	"fmt"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// root returns the root directory of the project
func root() *Directory {
	return dag.CurrentModule().Source().Directory(filepath.Join("..", "..", ".."))
}

func GoVersion(ctx context.Context, gomod *File) (string, error) {
	mod, err := gomod.Contents(ctx)
	if err != nil {
		return "", err
	}

	f, err := modfile.Parse("go.mod", []byte(mod), nil)
	if err != nil {
		return "", err
	}

	return f.Go.Version, nil
}

func modCache(container *Container) *Container {
	return container.WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache"))
}

func buildCache(container *Container) *Container {
	return container.WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build-cache"))
}

func goBase(version string) *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s", version)).
		With(modCache).
		With(buildCache)
}
