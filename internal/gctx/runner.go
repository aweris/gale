package gctx

import "os"

type RunnerContext struct {
	Name      string `json:"name" env:"RUNNER_NAME"`             // Name is the name of the runner.
	OS        string `json:"os" env:"RUNNER_OS"`                 // OS is the operating system of the runner.
	Arch      string `json:"arch" env:"RUNNER_ARCH"`             // Arch is the architecture of the runner.
	Temp      string `json:"temp" env:"RUNNER_TEMP"`             // Temp is the path to the directory containing temporary files created by the runner during the job.
	ToolCache string `json:"tool_cache" env:"RUNNER_TOOL_CACHE"` // ToolCache is the path to the directory containing installed tools.
	Debug     string `json:"debug" env:"RUNNER_DEBUG"`           // Debug is a boolean value that indicates whether to run the runner in debug mode.
}

func (c *Context) LoadRunnerContext() error {
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

	os.Setenv("RUNNER_NAME", c.Runner.Name)
	os.Setenv("RUNNER_OS", c.Runner.OS)
	os.Setenv("RUNNER_ARCH", c.Runner.Arch)
	os.Setenv("RUNNER_TEMP", c.Runner.Temp)
	os.Setenv("RUNNER_TOOL_CACHE", c.Runner.ToolCache)
	os.Setenv("RUNNER_DEBUG", c.Runner.Debug)

	return nil
}
