package preflight

import (
	"context"

	"github.com/aweris/gale/internal/core"
)

// Options represents the options for the preflight checks.
type Options struct {
	Repo         string // Repo is the name of the repository. It should be in the format of owner/repo.
	Branch       string // Branch is the name of the branch.
	Tag          string // Tag is the name of the tag.
	WorkflowsDir string // WorkflowsDir is the directory that contains the workflows.
	Workflow     string // Workflow is the name of the workflow.
	Job          string // Job is the name of the job
}

// Context represents the context of the preflight checks.
type Context struct {
	Context       context.Context               // Context is the context of the preflight checks.
	Repo          *core.Repository              // Repo represents a GitHub repository that is used for the preflight checks.
	Workflows     map[string]*core.Workflow     // Workflows is the list of workflows that are loaded.
	Workflow      *core.Workflow                // Workflow is the workflow that is used for the preflight checks.
	Job           *core.Job                     // Job is the job that is used for the preflight checks.
	CustomActions map[string]*core.CustomAction // CustomActions is the list of custom actions that are loaded.
	Shells        map[string]bool               // Shells is the list of shells that are used in the workflow.
	DockerImages  map[string]bool               // DockerImages is the list of Docker images that are used in the workflow.
}

// MessageLevel is the level of the message. It can be INFO, WARNING, or ERROR.
type MessageLevel string

const (
	Debug   MessageLevel = "DEBUG"
	Info    MessageLevel = "INFO"
	Warning MessageLevel = "WARNING"
	Error   MessageLevel = "ERROR"
)

// Message contains the level and the content of the message.
type Message struct {
	Level   MessageLevel // Level is the level of the message.
	Content string       // Content is the content of the message.
}

// ResultStatus is the status of the executed check. It can be PASSED or FAILED.
type ResultStatus string

const (
	Passed ResultStatus = "PASSED" // Passed represents a successful check.
	Failed ResultStatus = "FAILED" // Failed represents a failed check.
)

// Result contains the status of the check and the messages.
type Result struct {
	Status   ResultStatus // Status is the status of the check.
	Messages []Message    // Messages are the messages returned by the check.
}
