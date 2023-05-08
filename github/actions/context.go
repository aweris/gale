package actions

import "os"

// RunContext represents the root context of github actions run.
type RunContext struct {
	Github *GithubContext
	Runner *RunnerContext
}

func (r *RunContext) ToEnv() Environment {
	env := Environment{}
	env = env.Merge(r.Github.ToEnv())
	env = env.Merge(r.Runner.ToEnv())
	return env
}

func NewDummyContext() *RunContext {
	return &RunContext{
		Github: &GithubContext{
			Repository: "dagger/dagger",
			Workspace:  "/home/runner/work/dagger/dagger",
			Token:      os.Getenv("GITHUB_TOKEN"),
		},
		Runner: NewRunnerContext(false),
	}
}
