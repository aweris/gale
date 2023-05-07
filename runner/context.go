package runner

import (
	"dagger.io/dagger"
	"github.com/aweris/gale/github/actions"
	"github.com/aweris/gale/internal/event"
	"github.com/aweris/gale/logger"
)

var _ event.Context = new(Context)

type Context struct {
	client    *dagger.Client
	container *dagger.Container

	workflow *actions.Workflow
	job      *actions.Job
	context  *actions.RunContext

	stepResults         map[string]*actions.StepResult
	stepState           map[string]map[string]string
	actionsBySource     map[string]*actions.Action
	actionPathsBySource map[string]string

	log logger.Logger
}

// NewContext creates a minimal context for a runner
func NewContext(client *dagger.Client, log logger.Logger) *Context {
	return &Context{
		client:              client,
		stepResults:         make(map[string]*actions.StepResult),
		stepState:           make(map[string]map[string]string),
		actionsBySource:     make(map[string]*actions.Action),
		actionPathsBySource: make(map[string]string),
		log:                 log,
	}
}
