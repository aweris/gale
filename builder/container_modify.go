package builder

import (
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

// ModifyFn is a function that modifies a Dagger container.
type ModifyFn func(container *dagger.Container) *dagger.Container

// WithStep adds a step to the builder. Steps are executed in order.
func (b *Builder) WithStep(step ModifyFn) *Builder {
	b.modifyFns = append(b.modifyFns, step)
	return b
}

// WithCombinedExec adds a step to the builder that executes the given commands in a single step. This is useful for
// chaining multiple commands together and reducing the number of layers in the resulting image.
func (b *Builder) WithCombinedExec(commands ...string) *Builder {
	return b.WithStep(
		func(container *dagger.Container) *dagger.Container {
			return container.WithExec([]string{"sh", "-c", strings.Join(commands, " && ")})
		},
	)
}

// WithAptPackages adds steps to the builder that install the given apt packages. The package list is updated and
// removed in the same step to reduce the number of layers in the resulting image.
func (b *Builder) WithAptPackages(packages ...AptPackage) *Builder {
	install := make([]string, 0, len(packages))

	for _, p := range packages {
		install = append(install, p.String())
	}

	// to make sure we have some packages to install
	if len(install) == 0 {
		return b
	}

	// update the package list and install the packages and remove the package list in same step to reduce
	// the number of layers
	b.WithCombinedExec(
		"apt-get update -y",
		fmt.Sprintf("apt-get install -y --no-install-recommends %s", strings.Join(install, " ")),
		"rm -rf /var/lib/apt/lists/*",
	)

	return b
}

// WithToolSets installs the given tools using the original runner install script. The scripts are mounted from the
// original runner images repo.
//
// See: https://github.com/actions/runner-images/tree/main/images/linux/scripts/installers
func (b *Builder) WithToolSets(tools ...string) *Builder {
	for _, tool := range tools {
		b.WithStep(func(container *dagger.Container) *dagger.Container {
			return container.WithExec([]string{"bash", "/tmp/runner/toolset/installers/" + tool + ".sh"})
		})
	}
	return b
}

// WithBuildConfig applies the given build config to the builder.
func (b *Builder) WithBuildConfig(config *Config) *Builder {
	b.WithRunnerLabel(config.Label)

	// To improve readability, instead of using if-else statements, we use multiple if statements and allow the
	// last one to override the previous ones.

	if config.From != "" {
		b.From(config.From)
	}

	if config.DockerFile != "" {
		if filepath.IsAbs(config.DockerFile) {
			dir := b.client.Host().Directory(filepath.Dir(config.DockerFile))
			dockerfile := filepath.Base(config.DockerFile)
			b.Dockerfile(dir, dagger.ContainerBuildOpts{Dockerfile: dockerfile})
		} else {
			b.Dockerfile(b.client.Host().Directory("."), dagger.ContainerBuildOpts{Dockerfile: config.DockerFile})
		}
	}

	if config.ImportPath != "" {
		file := b.client.Host().Directory(filepath.Dir(config.ImportPath)).File(filepath.Base(config.ImportPath))
		b.Import(file)
	}

	// To improve readability, instead of using if-else statements, we use multiple if statements and use continue
	// to skip the rest of the statements.

	for _, modification := range config.Modifications {
		if modification.Run != "" {
			b.WithCombinedExec(modification.Run)
			continue
		}

		if modification.AptPackages != nil {
			b.WithAptPackages(modification.AptPackages...)
			continue
		}

		if modification.Tools != nil {
			b.WithToolSets(modification.Tools...)
			continue
		}
	}

	return b
}
