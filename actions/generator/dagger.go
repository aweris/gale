package main

import (
	"fmt"
	"strings"
)

type DaggerCli struct {
	ctr *Container
}

func NewDaggerCli(version string) *DaggerCli {
	// if version is set, use it. Otherwise, omit it and default behavior is to use the latest version.
	if version != "" {
		// remove the leading v if it exists
		version = strings.TrimPrefix(version, "v")

		// prefix the version with DAGGER_VERSION=
		version = fmt.Sprintf("DAGGER_VERSION=%s", version)
	}

	ctr := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -L https://dl.dagger.io/dagger/install.sh | %s sh", version)}).
		WithEntrypoint([]string{"/bin/dagger"})

	return &DaggerCli{ctr: ctr}
}

func (m *DaggerCli) InitModule(name string, deps ...string) *Directory {
	// base container with dagger cli and working directory set to /module
	ctr := m.ctr.WithWorkdir("/module")

	// initialize the module and update go.mod
	ctr = ctr.WithExec([]string{"mod", "init", "--name", name, "--sdk", "go", "--silent"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	ctr = ctr.WithExec([]string{"sed", "-i", "s/module main/module " + name + "/", "go.mod"}, ContainerWithExecOpts{SkipEntrypoint: true})

	// add given dependencies to dagger json
	for _, dep := range deps {
		ctr = ctr.WithExec([]string{"mod", "use", dep}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	}

	// return the directory containing the initialized module
	return ctr.Directory("/module")
}
