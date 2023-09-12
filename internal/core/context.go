package core

import (
	"encoding/json"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

var _ helpers.WithContainerFuncHook = new(RunnerContext)

type RunnerContext struct {
	Name      string `json:"name"`       // Name is the name of the runner.
	OS        string `json:"os"`         // OS is the operating system of the runner.
	Arch      string `json:"arch"`       // Arch is the architecture of the runner.
	Temp      string `json:"temp"`       // Temp is the path to a temporary directory on the runner.
	ToolCache string `json:"tool_cache"` // ToolCache is the path to the directory containing preinstalled tools for GitHub-hosted runners.
	Debug     string `json:"debug"`      // Debug is set only if debug logging is enabled, and always has the value of 1.
}

// NewRunnerContext creates a new RunnerContext from the given runner.
func NewRunnerContext(debug bool) RunnerContext {
	// adjust debug value from bool to string as it's expected to be string in the container.
	debugVal := "0"
	if debug {
		debugVal = "1"
	}
	return RunnerContext{
		Name:      "Gale Agent",
		OS:        "linux",
		Arch:      "x64",
		Temp:      "/home/runner/_temp",
		ToolCache: "/home/runner/hostedtoolcache", // /opt/hostedtoolcache is used by our base runner image and if we mount it we'll lose the tools installed by the base image.
		Debug:     debugVal,
	}
}

func (c RunnerContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithEnvVariable("RUNNER_NAME", c.Name).
			WithEnvVariable("RUNNER_TEMP", c.Temp).
			WithEnvVariable("RUNNER_OS", c.OS).
			WithEnvVariable("RUNNER_ARCH", c.Arch).
			WithEnvVariable("RUNNER_TOOL_CACHE", c.ToolCache).
			WithMountedCache(c.ToolCache, config.Client().CacheVolume("RUNNER_TOOL_CACHE")).
			WithEnvVariable("RUNNER_DEBUG", c.Debug)
	}
}

// JobContext contains information about the currently running job.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
type JobContext struct {
	Status Conclusion `json:"status"` // Status is the current status of the job. Possible values are success, failure, or cancelled.

	// TODO: add other fields when needed.
}

// StepContext is a context that contains information about the step.
//
// This context created per step execution. It won't be used or applied to the container level.
type StepContext struct {
	Conclusion Conclusion        `json:"conclusion"` // Conclusion is the result of a completed step after continue-on-error is applied
	Outcome    Conclusion        `json:"outcome"`    // Outcome is  the result of a completed step before continue-on-error is applied
	Outputs    map[string]string `json:"outputs"`    // Outputs is a map of output name to output value
	State      map[string]string `json:"-"`          // State is a map of step state variables. This is not available to expressions so that's why json tag is set to "-" to ignore it.
	Summary    string            `json:"-"`          // Summary is the summary of the step. This is not available to expressions so that's why json tag is set to "-" to ignore it.
}

var _ helpers.WithContainerFuncHook = new(SecretsContext)

// SecretsContext is a context that contains secrets.
type SecretsContext map[string]string

// NewSecretsContext creates a new SecretsContext from the given secrets.
func NewSecretsContext(token string, secrets map[string]string) SecretsContext {
	if secrets == nil {
		secrets = make(map[string]string)
	}

	secrets["GITHUB_TOKEN"] = token // GITHUB_TOKEN is a special secret that is always available to the workflow.

	return secrets
}

func (c SecretsContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		data, err := json.Marshal(c)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		secret := config.Client().SetSecret("secrets-context", string(data))

		return container.WithMountedSecret(filepath.Join(config.GhxHome(), "secrets", "secrets.json"), secret)
	}
}
