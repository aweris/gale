package gale

import (
	"context"
	"encoding/json"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/pkg/model"
	"github.com/aweris/gale/pkg/repository"
)

// containerRunnerPath path where the state and artifacts are stored in the container
const containerRunnerPath = "/home/runner/_temp/ghx"

// runnerCmdContainerPath is the path to the runner binary in the container
const runnerCmdContainerPath = "/usr/local/bin/ghx"

type ModifierFn func(container *dagger.Container) (*dagger.Container, error)

// Gale is the main struct for the gale package
type Gale struct {
	client      *dagger.Client
	base        *dagger.Container
	repo        *repository.Repo
	modifierFns []ModifierFn
}

// New creates a new gale instance
func New(client *dagger.Client) *Gale {
	return &Gale{client: client, modifierFns: []ModifierFn{initContainer(client)}}
}

func NewFromContainer(client *dagger.Client, base *dagger.Container) *Gale {
	return &Gale{client: client, base: base, modifierFns: []ModifierFn{initContainer(client)}}
}

// WithRepository sets the repository for the gale instance. This is optional. If not set, the current repository will
// be used.
func (g *Gale) WithRepository(repo *repository.Repo) *Gale {
	g.repo = repo // just to keep the repo in the gale instance to be used later.

	g.WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
		repoPath := filepath.Join("/home/runner/work", repo.Name, repo.Name)
		repoDir := g.client.Host().Directory(repo.Path)

		container = container.WithDirectory(repoPath, repoDir)

		// just to make sure that the github workspace is set to the repo path. This is required by some actions.
		// TODO: make sure this is the correct way to do this. It'll be possibly overwritten with different value by the github context
		container = container.WithEnvVariable("GITHUB_WORKSPACE", repoPath)
		container = container.WithWorkdir(repoPath)

		return container, nil
	})
	return g
}

// WithJob sets workflow and job environment variables and configures to all steps in the job. This step requires the
// repository to be set and workflow and job to be defined in the repository.
func (g *Gale) WithJob(workflow, job string) *Gale {
	args := []string{"with", "job", "--workflow=" + workflow, "--job=" + job}

	g.WithModifier(
		func(container *dagger.Container) (*dagger.Container, error) {
			return container.WithExec(getRunnerCommandWithArgs(args...)), nil
		},
	)

	return g
}

// TODO: make override optional argument

// WithStep adds a step to the steps configuration file to be executed by the runner
func (g *Gale) WithStep(step *model.Step, override bool) *Gale {
	// TODO: probably we can pass step as json/yaml to the runner and let the runner parse it.
	args := []string{"with", "step"}

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

	g.WithModifier(
		func(container *dagger.Container) (*dagger.Container, error) {
			return container.WithExec(getRunnerCommandWithArgs(args...)), nil
		},
	)

	return g
}

// WithGithubContext sets the github and runner contexts for the gale instance.
func (g *Gale) WithGithubContext(github *model.GithubContext, runner *model.RunnerContext) *Gale {
	return g.WithModifier(func(container *dagger.Container) (*dagger.Container, error) {
		for k, v := range github.ToEnv() {
			container = container.WithEnvVariable(k, v)
		}

		for k, v := range runner.ToEnv() {
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

		// make sure directory for GITHUB_PATH exists
		container = container.WithDirectory(filepath.Dir(github.Path), g.client.Host().Directory(filepath.Dir(github.Path)))

		return container, nil
	})
}

// WithModifier adds a modifier function to the gale instance.
func (g *Gale) WithModifier(fn ModifierFn) *Gale {
	g.modifierFns = append(g.modifierFns, fn)
	return g
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

type ExecResult struct {
	Container *dagger.Container
}

// ExportRunnerDirectory exports the runner directory contains all configuration, logs and artifacts to the host
// machine. This is useful for debugging purposes.
func (r *ExecResult) ExportRunnerDirectory(ctx context.Context, path string) error {
	_, err := r.Container.Directory(containerRunnerPath).Export(ctx, path)
	if err != nil {
		return err
	}

	return nil
}

func (g *Gale) Exec(ctx context.Context) (*ExecResult, error) {
	container, err := g.Container()
	if err != nil {
		return nil, err
	}

	result := &ExecResult{Container: container}

	result.Container = result.Container.WithExec(getRunnerCommandWithArgs("run"))

	_, err = result.Container.ExitCode(ctx)
	if err != nil {
		return result, err
	}

	return result, nil
}

// initContainer initializes the container with the necessary files and directories. If the container is nil, a new
// container will be created. This modifier function should be called before any other modifier functions.
func initContainer(client *dagger.Client) ModifierFn {
	return func(container *dagger.Container) (*dagger.Container, error) {
		if container == nil {
			container = client.Container().From("ghcr.io/catthehacker/ubuntu:act-22.04")
		}

		runnerFile := buildExecutorFromSource(client)

		container = container.WithUnixSocket("/var/run/docker.sock", client.Host().UnixSocket("/var/run/docker.sock"))
		container = container.WithFile(runnerCmdContainerPath, runnerFile, dagger.ContainerWithFileOpts{Permissions: 0755})

		return container, nil
	}
}

func getRunnerCommandWithArgs(args ...string) []string {
	return append([]string{runnerCmdContainerPath}, args...)
}
