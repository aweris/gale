package runner

import (
	"context"
	"dagger.io/dagger"
	"embed"
)

// runnerScripts contains the scripts used to initialize the runner. These scripts are used for keeping provisioning
// logic in one place and simplifying the code.
//
//go:embed scripts/*.sh
var runnerScripts embed.FS

// Runner represents a GitHub Action runner powered by Dagger.
type Runner struct {
	// Container is the Dagger container that the runner is running in.
	Container *dagger.Container
}

// NewRunner creates a new Runner.
func NewRunner(ctx context.Context, client *dagger.Client) (*Runner, error) {
	var (
		err       error
		container *dagger.Container
	)

	container = client.Container().From("docker.io/library/ubuntu:22.04")

	container, err = initRunner(container)
	if err != nil {
		return nil, err
	}

	container = installTools(client, container)

	// run as non-root user runner like the original runner image
	container = container.WithUser("runner")

	// provision the container before returning in order to fail early if there are any issues
	_, err = container.ExitCode(ctx)
	if err != nil {
		return nil, err
	}

	return &Runner{Container: container}, nil
}

// initRunner initializes the runner.
func initRunner(container *dagger.Container) (*dagger.Container, error) {
	// To keep initial provisioning simple, we use the same scripts to initialize the container.
	// TODO: improve provisioning logic, using scripts is not the best way to do this. It hides the logic and makes it
	//  hard to debug.
	b, err := runnerScripts.ReadFile("scripts/init.sh")
	if err != nil {
		return nil, err
	}

	var (
		path = "/tmp/runner/scripts/init.sh"
		opts = dagger.ContainerWithNewFileOpts{Contents: string(b), Permissions: 0755}
	)

	container = container.WithNewFile(path, opts)
	container = container.WithExec([]string{path})

	return container, nil
}

// installTools installs the tools required by the runner.
func installTools(client *dagger.Client, container *dagger.Container) *dagger.Container {
	// original runner images repo. We try to use same scripts and toolset.json to keep things consistent
	// original repo.
	toolsetRepo := client.Git("https://github.com/actions/runner-images.git").Branch("main").Tree()

	var (
		// toolset json for ubuntu 22.04
		toolsetJson = toolsetRepo.File("images/linux/toolsets/toolset-2204.json")

		// directory containing helper scripts
		helperScriptsDir = toolsetRepo.Directory("images/linux/scripts/helpers")

		// directory containing installer scripts
		installersScriptsDir = toolsetRepo.Directory("images/linux/scripts/installers")
	)

	// /imagegeneration/installers/toolset.json is defined in install helper script statically
	container = container.WithFile("/imagegeneration/installers/toolset.json", toolsetJson)

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
}
