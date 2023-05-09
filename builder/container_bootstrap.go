package builder

import (
	"fmt"
	"path/filepath"
	"strconv"

	"dagger.io/dagger"

	"github.com/aweris/gale/config"
)

const toolsetPath = "/tmp/runner/toolset"

// bootstrap adds the default steps to the builder. This includes installing dependencies, creating the runner user,
// mounting the toolset repository and installing the default toolsets. This function is called by NewBuilder.
func (b *Builder) bootstrap() *Builder {
	// Set default container creation function
	b.From(config.DefaultRunnerImage)

	// Add default modifyFns to the builder
	b.installDependencies()
	b.createGroup(121, "docker")
	b.createUser(1001, "runner", "docker", "sudo")
	b.mountToolsetRepo()
	b.WithToolSets("nodejs")

	return b
}

// installDependencies adds a step to the builder that installs the dependencies required by the runner.
func (b *Builder) installDependencies() *Builder {
	return b.WithCombinedExec(
		"apt-get update -y",
		"apt-get install -y software-properties-common",
		"add-apt-repository -y ppa:git-core/ppa",
		"apt-get update -y",
		"apt-get install -y --no-install-recommends git curl ca-certificates jq sudo unzip zip strace",
		"rm -rf /var/lib/apt/lists/*",
	)
}

// createGroup adds a step to the builder that creates a group with the given GID and name.
func (b *Builder) createGroup(gid int, group string) *Builder {
	return b.WithStep(
		func(container *dagger.Container) *dagger.Container {
			return container.WithExec([]string{"groupadd", "--force", "--gid", strconv.Itoa(gid), group})
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

	// allocate just enough space for the commands we need
	commands := make([]string, 0, 8)

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
	// issue, we createFn the directory and set owner to runner.
	commands = append(commands, "mkdir -p "+tempDir)
	commands = append(commands, "chown -R "+user+":"+user+" "+tempDir)

	return b.WithCombinedExec(commands...)
}

// TODO: read this from release and create list of tools to install from file names
func (b *Builder) mountToolsetRepo() *Builder {
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
			container = container.WithDirectory(filepath.Join(toolsetPath, "helpers"), helperScriptsDir)
			container = container.WithEnvVariable("HELPER_SCRIPTS", "/tmp/runner/toolset/helpers")

			// add dummy scripts to container instead of using the original script. This will allow us to use original
			// scripts without any modification.
			container = container.WithNewFile("/usr/local/bin/invoke_tests", dagger.ContainerWithNewFileOpts{Contents: "echo \"$@\"", Permissions: 0755})
			container = container.WithNewFile("/usr/local/bin/systemctl", dagger.ContainerWithNewFileOpts{Contents: "echo \"$@\"", Permissions: 0755})

			// installers scripts are used to install tools
			container = container.WithDirectory(filepath.Join(toolsetPath, "installers"), installersScriptsDir, dagger.ContainerWithDirectoryOpts{Owner: "runner:runner"})

			return container
		},
	)
}
