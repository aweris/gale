package model

import "os"

// DaggerContext represents the dagger engine connection information should be passed to the container
type DaggerContext struct {
	RunnerHost string // where the dagger engine is running
	Session    string // dagger session information
}

// NewDaggerContextFromEnv creates a new dagger context from environment variables
func NewDaggerContextFromEnv() *DaggerContext {
	return &DaggerContext{
		RunnerHost: os.Getenv("_EXPERIMENTAL_DAGGER_RUNNER_HOST"),
		Session:    os.Getenv("DAGGER_SESSION"),
	}
}

// ToEnv converts the dagger context to environment variables
func (d *DaggerContext) ToEnv() map[string]string {
	env := make(map[string]string)

	if d.RunnerHost != "" {
		env["_EXPERIMENTAL_DAGGER_RUNNER_HOST"] = d.RunnerHost
	}

	if d.Session != "" {
		env["DAGGER_SESSION"] = d.Session
	}

	return env
}
