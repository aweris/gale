package gale

import (
	"context"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/gctx"
)

// Gale is the main entrypoint for the gale library.
type Gale struct {
	rc  *gctx.Context
	ghx *Ghx
}

// New creates a new gale instance.
func New(rc *gctx.Context) *Gale {
	return &Gale{
		rc:  rc,
		ghx: NewGhxBinary(rc.Repo.CacheVol),
	}
}

// ExecutionEnv returns a dagger function that sets the execution environment of the gale to the given container.
func (g *Gale) ExecutionEnv(_ context.Context) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// context configuration
		container = container.With(g.rc.WithContainerFunc())

		// tools
		container = container.With(g.ghx.WithContainerFunc())

		return container
	}
}

// Run runs a job from a workflow.
func (g *Gale) Run(workflow, job string) dagger.WithContainerFunc {
	return g.ghx.Run(workflow, job)
}
