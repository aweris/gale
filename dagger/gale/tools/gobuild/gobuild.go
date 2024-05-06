// package gobuild provides a way to build monorepo binaries.

package gobuild

import (
	"context"

	"golang.org/x/mod/modfile"

	"dagger/gale/internal/dagger"
)

type options struct {
	include []string
	exclude []string
	workdir string
}

type Opt func(*options)

// Include sets the include option.
func Include(include ...string) Opt {
	return func(o *options) { o.include = include }
}

func Exclude(exclude ...string) Opt {
	return func(o *options) { o.exclude = exclude }
}

// Workdir sets the workdir option.
func Workdir(workdir string) Opt {
	return func(o *options) { o.workdir = workdir }
}

func Build(ctx context.Context, dag *dagger.Client, opts ...Opt) (*dagger.File, error) {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	var (
		code = dag.Directory().WithDirectory(".", root(dag), dagger.DirectoryWithDirectoryOpts{Include: o.include, Exclude: o.exclude})
		mod  = code.Directory(o.workdir).File("go.mod")
	)

	version, err := version(ctx, mod)
	if err != nil {
		return nil, err
	}

	return dag.Go(dagger.GoOpts{Version: version}).Container().
		WithMountedDirectory("/src", code).
		WithWorkdir("/src/" + o.workdir).
		WithExec([]string{"go", "mod", "download"}).
		WithExec([]string{"go", "build", "-o", "bin/cmd", "."}).
		File("bin/cmd"), nil
}

// returns the root directory of the monorepo.
func root(dag *dagger.Client) *dagger.Directory {
	return dag.CurrentModule().Source().Directory("../../..")
}

// parses the given module's go.mod file and returns the Go version.
func version(ctx context.Context, gomod *dagger.File) (string, error) {
	data, err := gomod.Contents(ctx)
	if err != nil {
		return "", err
	}

	f, err := modfile.Parse("go.mod", []byte(data), nil)
	if err != nil {
		return "", err
	}

	return f.Go.Version, nil
}
