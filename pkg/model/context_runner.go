package model

// RunnerContext contains information about the runner that is executing the current job.
// All fields in this section are based on the Runner context documentation. Not all of them meaningful
// for gale, but we include them all for completeness.
//
// See more: https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
type RunnerContext struct {
	// The name of the runner executing the job.
	Name string `json:"name"`

	// The operating system of the runner executing the job. Possible values are Linux, Windows, or macOS.
	OS string `json:"os"`

	// The architecture of the runner executing the job. Possible values are X86, X64, ARM, or ARM64.
	Arch string `json:"arch"`

	// The path to a temporary directory on the runner. This directory is emptied at the beginning and end of
	// each job. Note that files will not be removed if the runner's user account does not have permission to
	// delete them.
	Temp string `json:"temp"`

	// The path to the directory containing preinstalled tools for GitHub-hosted runners.
	ToolCache string `json:"tool_cache"`

	// This is set only if debug logging is enabled, and always has the value of 1. It can be useful as an
	// indicator to enable additional debugging or verbose logging in your own job steps.
	Debug string `json:"debug"`
}

func NewRunnerContext(debug bool) *RunnerContext {
	debugVal := ""
	if debug {
		debugVal = "1"
	}

	return &RunnerContext{
		Name:      "", // TODO: Not sure if we need this at all. Remove after double-checking.
		OS:        "linux",
		Arch:      "x64", // TODO: This should be determined by the host
		Temp:      "/home/runner/_temp",
		ToolCache: "/home/runner/_tool",
		Debug:     debugVal,
	}
}

// ToEnv converts the RunnerContext to a map of environment variables.
// More info: https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables
func (r *RunnerContext) ToEnv() map[string]string {
	return map[string]string{
		"RUNNER_NAME":       r.Name,
		"RUNNER_TEMP":       r.Temp,
		"RUNNER_OS":         r.OS,
		"RUNNER_ARCH":       r.Arch,
		"RUNNER_TOOL_CACHE": r.ToolCache,
		"RUNNER_DEBUG":      r.Debug,
	}
}
