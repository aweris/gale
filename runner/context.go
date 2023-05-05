package runner

import (
	"dagger.io/dagger"
	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/internal/event"
	"github.com/aweris/gale/logger"
)

var _ event.Context = new(Context)

type Context struct {
	client    *dagger.Client
	container *dagger.Container

	workflow *gha.Workflow
	job      *gha.Job
	context  *gha.RunContext

	stepResults         map[string]*gha.StepResult
	stepState           map[string]map[string]string
	actionsBySource     map[string]*gha.Action
	actionPathsBySource map[string]string

	log logger.Logger
}
