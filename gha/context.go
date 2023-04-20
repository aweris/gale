package gha

import "os"

// RunContext represents the root context of github actions run.
type RunContext struct {
	Github GithubContext
	Runner RunnerContext
}

func (r *RunContext) ToEnv() Environment {
	env := Environment{}
	env = env.Merge(r.Github.ToEnv())
	env = env.Merge(r.Runner.ToEnv())
	return env
}

// GithubContext contains information about the workflow run and the event that triggered the run.
type GithubContext struct {
	Repository string
	Workspace  string
	Token      string
}

func (g *GithubContext) ToEnv() Environment {
	return Environment{
		"GITHUB_REPOSITORY": g.Repository,
		"GITHUB_WORKSPACE":  g.Workspace,
		"GITHUB_TOKEN":      g.Token,
	}
}

// RunnerContext contains information about the runner that is executing the current job.
type RunnerContext struct {
	Temp      string
	ToolCache string
}

func (r *RunnerContext) ToEnv() Environment {
	return Environment{
		"RUNNER_TEMP":       r.Temp,
		"RUNNER_TOOL_CACHE": r.ToolCache,
	}
}

func NewDummyContext() *RunContext {
	return &RunContext{
		Github: GithubContext{
			Repository: "dagger/dagger",
			Workspace:  "/home/runner/work/dagger/dagger",
			Token:      os.Getenv("GITHUB_TOKEN"),
		},
		Runner: RunnerContext{
			Temp:      "/home/runner/_temp",
			ToolCache: "/home/runner/_tool",
		},
	}
}
