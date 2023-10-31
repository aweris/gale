package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

const DefaultRuntimeRef = "c9b01b328a59ec6452eb451ebf0e9b2a1280a504"

// ActionsGenerator generates dagger modules using Github Actions.
type ActionsGenerator struct{}

func (m *ActionsGenerator) Generate(
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

	// FIXME: need to look for latest dagger version. Not static version.
	readme := generateModuleREADME(ca, daggerVersion.GetOr("v0.9.1"))

	var (
		runtime    = runtimeVersion.GetOr(DefaultRuntimeRef)
		runtimeDep = fmt.Sprintf("github.com/aweris/gale/daggerverse/actions/runtime@%s", runtime)
		dagger     = dagger(daggerVersion)
		opt        = ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}
		cmdModInit = []string{"mod", "init", "--name", filepath.Base(ca.Repo), "--sdk", "go", "--silent"}
		cmdModUse  = []string{"mod", "use", runtimeDep}
	)

	// create a new dagger module and replace the initial main.go with the generated main.go
	source := dagger.
		WithWorkdir("/module").
		WithExec(cmdModInit, opt).
		WithExec(cmdModUse, opt).
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
