package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"golang.org/x/mod/modfile"
)

// root returns the root directory of the project.
func root() string {
	// get location of current file
	_, current, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(current), "../..")
}

// root returns the root directory of the project.
func (h *Host) root(opts ...HostDirectoryOpts) *Directory {
	return h.Directory(root(), opts...)
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

func ModCache(container *Container) *Container {
	return container.WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod-cache"))
}

func BuildCache(container *Container) *Container {
	return container.WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build-cache"))
}

func GoBase(version string) *Container {
	return dag.Container().
		From(fmt.Sprintf("golang:%s", version)).
		With(ModCache).
		With(BuildCache)
}
