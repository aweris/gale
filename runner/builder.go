package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
)

// ModifyFn is a function that modifies a Dagger container.
type ModifyFn func(container *dagger.Container) *dagger.Container

// Builder represents a builder for creating a GitHub Action runner.
type Builder struct {
	// client is the Dagger client used to create the container.
	client *dagger.Client

	// label is the label of the container.
	label string

	// Docker image address container is created from
	from string

	// build steps to be executed in order to create the runner
	steps []ModifyFn
}

// NewBuilder creates a new Builder.
func NewBuilder(client *dagger.Client) *Builder {
	// create a new builder instance with default values
	builder := &Builder{
		client: client,
		from:   config.DefaultRunnerImage,
		label:  config.DefaultRunnerLabel,
	}

	// Add default steps to the builder
	builder.installDependencies()
	builder.createGroup(121, "docker")
	builder.createUser(1001, "runner", "docker", "sudo")
	builder.installTools()

	return builder
}

// WithRunnerLabel sets the label for the runner container.
func (b *Builder) WithRunnerLabel(label string) *Builder {
	b.label = label
	return b
}

// From sets the Docker image address container is created from.
func (b *Builder) From(from string) *Builder {
	b.from = from
	return b
}

// WithStep adds a step to the builder. Steps are executed in order.
func (b *Builder) WithStep(step ModifyFn) *Builder {
	b.steps = append(b.steps, step)
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

// installDependencies adds a step to the builder that installs the dependencies required by the runner.
func (b *Builder) installDependencies() *Builder {
	return b.WithCombinedExec(
		"apt-get update -y",
		"apt-get install -y software-properties-common",
		"add-apt-repository -y ppa:git-core/ppa",
		"apt-get update -y",
		"apt-get install -y --no-install-recommends git curl ca-certificates jq sudo unzip zip",
		"rm -rf /var/lib/apt/lists/*",
	)
}

// createGroup adds a step to the builder that creates a group with the given GID and name.
func (b *Builder) createGroup(gid int, group string) *Builder {
	return b.WithStep(
		func(container *dagger.Container) *dagger.Container {
			return container.WithExec([]string{"groupadd", "--gid", strconv.Itoa(gid), group})
		},
	)
}

// createUser adds a step to the builder that creates a user with the given UID and name.
func (b *Builder) createUser(uid int, user string, groups ...string) *Builder {
	var (
		homeDir    = fmt.Sprintf("/home/%s", user)
		tempDir    = fmt.Sprintf("/home/%s/_temp", user)
		actionsDir = "/home/actions"
	)

	var commands []string

	// Create new user
	commands = append(commands, "useradd --create-home --home-dir "+homeDir+" --uid "+strconv.Itoa(uid)+" "+user)

	// Add user to groups
	for _, group := range groups {
		commands = append(commands, "usermod -aG "+group+" "+user)
	}

	// Allow runner to run sudo without password
	commands = append(commands, "echo 'runner ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers")

	// Allow sudo to set DEBIAN_FRONTEND
	commands = append(commands, "echo 'Defaults env_keep += \"DEBIAN_FRONTEND\"' >> /etc/sudoers")

	// Create actions directory and set owner to runner home directory. This directory used by actions.
	commands = append(commands, "mkdir -p "+actionsDir)
	commands = append(commands, "chown -R "+user+":"+user+" "+actionsDir)

	// Temp directory is defined in RUNNER_TEMP environment variable. This directory will be used to store temporary files.
	// However, when we use dagger to mount directory to container, the directory will be owned by root. To workaround this
	// issue, we create the directory and set owner to runner.
	commands = append(commands, "mkdir -p "+tempDir)
	commands = append(commands, "chown -R "+user+":"+user+" "+tempDir)

	return b.WithCombinedExec(commands...)
}

// installTools installs the tools required by the runner.
func (b *Builder) installTools() *Builder {
	// original runner images repo. We try to use same scripts and toolset.json to keep things consistent
	// original repo.
	toolsetRepo := b.client.Git("https://github.com/actions/runner-images.git").Branch("main").Tree()

	var (
		// toolset json for ubuntu 22.04
		toolsetJSON = toolsetRepo.File("images/linux/toolsets/toolset-2204.json")

		// directory containing helper scripts
		helperScriptsDir = toolsetRepo.Directory("images/linux/scripts/helpers")

		// directory containing installer scripts
		installersScriptsDir = toolsetRepo.Directory("images/linux/scripts/installers")
	)

	return b.WithStep(
		func(container *dagger.Container) *dagger.Container {
			// /imagegeneration/installers/toolset.json is defined in install helper script statically
			container = container.WithFile("/imagegeneration/installers/toolset.json", toolsetJSON)

			// HELPER_SCRIPTS is used by installer scripts to find and source helper scripts
			container = container.WithDirectory("/tmp/runner/toolset/helpers", helperScriptsDir)
			container = container.WithEnvVariable("HELPER_SCRIPTS", "/tmp/runner/toolset/helpers")

			// add dummy invoke_tests script to container instead of using the original script.
			container = container.WithNewFile("/usr/local/bin/invoke_tests", dagger.ContainerWithNewFileOpts{Contents: "echo \"$@\"", Permissions: 0755})

			// installers scripts are used to install tools
			container = container.WithDirectory("/tmp/runner/toolset/installers", installersScriptsDir)

			// Install nodejs using the original runner install script
			container = container.WithExec([]string{"bash", "/tmp/runner/toolset/installers/nodejs.sh"})

			return container
		},
	)
}

// Build builds and exports the runner in the data home directory with the given label and returns the runner instance.
func (b *Builder) Build(ctx context.Context) (*Runner, error) {
	container := b.client.Container()

	// Create the container from the given image.
	// TODO: allow loading the container from a tarball or using a Dockerfile.
	container = container.From(b.from)

	// apply all steps to the container
	for _, step := range b.steps {
		container = step(container)
	}

	// Set the user to the runner user. User is created in default steps added in NewBuilder.
	container.WithUser("runner")

	dh := config.DataHome()

	if err := os.MkdirAll(filepath.Join(dh, b.label), 0755); err != nil {
		return nil, err
	}

	// Export the container to a tarball in the data home directory($XDG_DATA_HOME/gale/<runner-label>/image.tar).
	// This tarball will be used avoid rebuilding the runner image every time and reduce relying on cache.
	_, err := container.Export(ctx, filepath.Join(dh, b.label, config.DefaultRunnerImageTar))
	if err != nil {
		return nil, err
	}

	return &Runner{Container: container}, nil
}
