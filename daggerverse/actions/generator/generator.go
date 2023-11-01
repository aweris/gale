package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

const DefaultRuntimeRef = "da8e5dada0126583ef9201b8d6b494af765b7d31"

// ActionsGenerator generates dagger modules using Github Actions.
type ActionsGenerator struct{}

func (m *ActionsGenerator) Generate(
	ctx context.Context,
	// The Github Actions repository to generate dagger modules for. Format: <action-repo>@<version>
	action string,
	// The actions/runtime version to use. If not specified, the default version will be used.
	runtimeVersion Optional[string],
	// The dagger version to use. If not specified, the latest version will be used.
	daggerVersion Optional[string],
) (*Directory, error) {
	// get custom action metadata from action.yml/action.yaml
	ca, err := getCustomAction(action)
	if err != nil {
		return nil, err
	}

	// generate main.go for the custom action module
	main, err := generateModuleCode(ca)
	if err != nil {
		return nil, err
	}

	version, err := latestDaggerVersion(ctx)
	if err != nil {
		return nil, err
	}

	// FIXME: need to look for latest dagger version. Not static version.
	readme := generateModuleREADME(ca, daggerVersion.GetOr(version))

	var (
		runtime     = runtimeVersion.GetOr(DefaultRuntimeRef)
		runtimeDep  = fmt.Sprintf("github.com/aweris/gale/daggerverse/actions/runtime@%s", runtime)
		dagger      = dagger(daggerVersion)
		opt         = ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}
		moduleName  = filepath.Base(ca.Repo)
		cmdModInit  = []string{"mod", "init", "--name", moduleName, "--sdk", "go", "--silent"}
		cmdModUse   = []string{"mod", "use", runtimeDep}
		cmdFixGoMod = []string{"sed", "-i", "s/module main/module " + moduleName + "/", "go.mod"} // replace module main with module <module-name> in go.mod
	)

	// create a new dagger module and replace the initial main.go with the generated main.go
	source := dagger.
		WithWorkdir("/module").
		WithExec(cmdModInit, opt).
		WithExec(cmdModUse, opt).
		WithExec(cmdFixGoMod, ContainerWithExecOpts{SkipEntrypoint: true}).
		Directory("/module").
		WithoutFile("main.go").                                                                // remove initial main.go from the directory
		WithFile("main.go", main).                                                             // add generated main.go
		WithFile("README.md", readme).                                                         // add generated README.md
		WithNewFile(".gitignore", "/dagger.gen.go\n/internal/querybuilder/\n/querybuilder/\n") // add generated .gitignore

	// FIXME: need to look why .gitignore is not being added to the module. For now, we are generating it manually.

	// return a directory prefixed with the owner/repo name to simplify extraction of the module to host filesystem
	return dag.Directory().WithDirectory(ca.Repo, source, DirectoryWithDirectoryOpts{Include: []string{"**/*", ".git*"}}), nil
}

// dagger returns a dagger container with the specified dagger version. If no version is specified, the latest version
// will be used.
func dagger(daggerVersion Optional[string]) *Container {
	version, ok := daggerVersion.Get()

	if ok {
		// remove the leading v if it exists
		version = strings.TrimPrefix(version, "v")

		// prefix the version with DAGGER_VERSION=
		version = fmt.Sprintf("DAGGER_VERSION=%s", version)
	}

	return dag.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -L https://dl.dagger.io/dagger/install.sh | %s sh", version)}).
		WithEntrypoint([]string{"/bin/dagger"})
}

// latestDaggerVersion returns the latest dagger version.
func latestDaggerVersion(ctx context.Context) (string, error) {
	versions, err := dag.Container().From("alpine/git:latest").
		WithExec([]string{"ls-remote", "--tags", "--sort=-v:refname", "https://github.com/dagger/dagger.git"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}

	// get the first line
	version := strings.Split(versions, "\n")[0]

	// get second column from the line
	version = strings.Split(version, "\t")[1]

	// remove refs/tags/ from the version
	version = strings.TrimPrefix(version, "refs/tags/")

	return version, nil
}
