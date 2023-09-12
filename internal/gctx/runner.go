package gctx

import (
	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

type RunnerContext struct {
	Name      string `json:"name" env:"RUNNER_NAME" container_env:"true"`             // Name is the name of the runner.
	OS        string `json:"os" env:"RUNNER_OS" container_env:"true"`                 // OS is the operating system of the runner.
	Arch      string `json:"arch" env:"RUNNER_ARCH" container_env:"true"`             // Arch is the architecture of the runner.
	Temp      string `json:"temp" env:"RUNNER_TEMP" container_env:"true"`             // Temp is the path to the directory containing temporary files created by the runner during the job.
	ToolCache string `json:"tool_cache" env:"RUNNER_TOOL_CACHE" container_env:"true"` // ToolCache is the path to the directory containing installed tools.
	Debug     string `json:"debug" env:"RUNNER_DEBUG" container_env:"true"`           // Debug is a boolean value that indicates whether to run the runner in debug mode.
}

func (c *Context) LoadRunnerContext() error {
	// if the workflow is running in a container, load the context from the environment variables.
	if c.isContainer {
		runner, err := NewContextFromEnv[RunnerContext]()
		if err != nil {
			return err
		}

		c.Runner = runner

		return nil
	}

	// otherwise, create a new context from static values.

	debug := "0"
	if c.debug {
		debug = "1"
	}

	c.Runner = RunnerContext{
		Name:      "Gale Agent",
		OS:        "linux",
		Arch:      "x64",
		Temp:      "/home/runner/_temp",
		ToolCache: "/home/runner/hostedtoolcache", // /opt/hostedtoolcache is used by our base runner image and if we mount it we'll lose the tools installed by the base image.
		Debug:     debug,
	}

	return nil
}

// helpers.WithContainerFuncHook interface to be loaded in the container.

var _ helpers.WithContainerFuncHook = new(RunnerContext)

func (c *RunnerContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// Load context in the container as environment variables or secrets.
		container = container.With(WithContainerEnv(config.Client(), c))

		// Apply extra container configuration
		container = container.WithMountedCache(c.ToolCache, config.Client().CacheVolume("RUNNER_TOOL_CACHE"))

		return container
	}
}
