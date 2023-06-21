package gale

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/dagger/images"
	"github.com/aweris/gale/pkg/gh"
	"github.com/aweris/gale/pkg/model"
)

const (
	// containerRunnerPath path where the state and artifacts are stored in the container
	containerRunnerPath = "/home/runner/_temp/ghx"

	// runnerExitCodeFile is the name of the file where the exit code of the runner is stored. Actual exit code is
	// written to this file by the command in the container.
	runnerExitCodeFile = "exit-code"
)

// ModifierFn is a function that allows to modify the underlying dagger container.
type ModifierFn func(container *dagger.Container) (*dagger.Container, error)

// Gale is the entry point for the Gale library
type Gale struct {
	client      *dagger.Client
	base        *dagger.Container
	modifierFns []ModifierFn
}

// New creates a new Gale instance
func New(client *dagger.Client) *Gale {
	return NewFromContainer(client, nil)
}

// NewFromContainer creates a new Gale instance from an existing container.
func NewFromContainer(client *dagger.Client, base *dagger.Container) *Gale {
	gale := &Gale{client: client, base: base}

	// adds the default modifier functions to the gale instance
	gale.init()
	gale.loadCurrentRepository()

	return gale
}

// WithModifier adds a modifier function to the gale instance.
func (g *Gale) WithModifier(fn ModifierFn) *Gale {
	g.modifierFns = append(g.modifierFns, fn)
	return g
}

// init initializes the container with the default configuration.
func (g *Gale) init() *Gale {
	return g.WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
		if container == nil {
			container = images.RunnerBase(g.client)
		}

		// check if _EXPERIMENTAL_DAGGER_RUNNER_HOST exists and if so, use it
		if val := os.Getenv("_EXPERIMENTAL_DAGGER_RUNNER_HOST"); val != "" {
			if strings.HasPrefix(val, "unix://") {
				socket := strings.TrimPrefix(val, "unix://")
				container = container.WithUnixSocket(socket, g.client.Host().UnixSocket(socket))
			}

			container = container.WithEnvVariable("_EXPERIMENTAL_DAGGER_RUNNER_HOST", val)
		}

		// check if DAGGER_SESSION exists and if so, use it
		if val := os.Getenv("DAGGER_SESSION"); val != "" {
			container = container.WithEnvVariable("DAGGER_SESSION", val)
		}

		// TODO: make this optional. If _EXPERIMENTAL_DAGGER_RUNNER_HOST is set, we don't need to mount the docker socket

		// check if DOCKER_HOST should overrides the default docker socket location
		hostDockerSocket := "/var/run/docker.sock"
		if dockerHost := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dockerHost, "unix://") {
			hostDockerSocket = strings.TrimPrefix(dockerHost, "unix://")
		}

		container = container.WithUnixSocket("/var/run/docker.sock", g.client.Host().UnixSocket(hostDockerSocket))
		container = container.WithFile("/usr/local/bin/ghx", withGHX(g.client, "v0.0.2"))

		// load the runner context into the container.
		for k, v := range model.NewRunnerContextFromEnv().ToEnv() {
			container = container.WithEnvVariable(k, v)
		}

		return container, nil
	})
}

// loadCurrentRepository loads the current repository into the container. This method uses github cli to get the
// current repository.
func (g *Gale) loadCurrentRepository() *Gale {
	return g.WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
		repo, err := gh.CurrentRepository()
		if err != nil {
			return nil, err
		}

		path := filepath.Join("/home/runner/work", repo.Name, repo.Name)
		dir := g.client.Host().Directory(".")

		container = container.WithDirectory(path, dir)

		// just to make sure that the github workspace is set to the repo path. This is required by some actions.
		// TODO: make sure this is the correct way to do this. It'll be possibly overwritten with different value by the github context
		container = container.WithEnvVariable("GITHUB_WORKSPACE", path)
		container = container.WithWorkdir(path)

		return container, nil
	})
}

// WithGithubContext sets the github and runner contexts for the gale instance.
func (g *Gale) WithGithubContext(github *model.GithubContext) *Gale {
	return g.WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
		for k, v := range github.ToEnv() {
			container = container.WithEnvVariable(k, v)
		}

		data, err := json.Marshal(github.Event)
		if err != nil {
			return nil, err
		}

		if len(data) == 0 {
			data = []byte("{}")
		}

		fileOpts := dagger.ContainerWithNewFileOpts{Contents: string(data), Permissions: 0644}
		container = container.WithNewFile(github.EventPath, fileOpts)

		return container, nil
	})
}

// WithJob sets workflow and job environment variables and configures to all steps in the job. This step requires the
// repository to be set and workflow and job to be defined in the repository.
func (g *Gale) WithJob(workflow, job string) *Gale {
	return g.WithModifier(
		func(container *dagger.Container) (*dagger.Container, error) {
			return container.WithExec([]string{"ghx", "with", "job", "--workflow=" + workflow, "--job=" + job}), nil
		},
	)
}

// TODO: make override optional argument

// WithStep adds a step to the steps configuration file to be executed by the runner
func (g *Gale) WithStep(step *model.Step, override bool) *Gale {
	// TODO: probably we can pass step as json/yaml to the runner and let the runner parse it.
	args := []string{"ghx", "with", "step"}

	if step.ID != "" {
		args = append(args, "--id="+step.ID)
	}

	if step.Name != "" {
		args = append(args, "--name="+step.Name)
	}

	if step.Run != "" {
		args = append(args, "--run="+step.Run)
	}

	if step.Shell != "" {
		args = append(args, "--shell="+step.Shell)
	}

	if step.Uses != "" {
		args = append(args, "--uses="+step.Uses)
	}

	for k, v := range step.Environment {
		args = append(args, "--env="+k+"="+v)
	}

	for k, v := range step.With {
		args = append(args, "--with="+k+"="+v)
	}

	if override {
		args = append(args, "--override")
	}

	return g.WithModifier(
		func(container *dagger.Container) (*dagger.Container, error) {
			return container.WithExec(args), nil
		},
	)
}

// Container returns the a dagger container after applying all the modifier functions.
func (g *Gale) Container() (container *dagger.Container, err error) {
	container = g.base

	for _, fn := range g.modifierFns {
		container, err = fn(container)
		if err != nil {
			return nil, err
		}
	}

	return container, nil
}

func withGHX(client *dagger.Client, version string) *dagger.File {
	return client.Container().
		From(fmt.Sprintf("ghcr.io/aweris/ghx:%s", version)).
		Directory("/usr/local/bin").
		File("ghx")
}
