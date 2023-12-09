package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/aweris/gale/common/log"
)

func New(
	// The context of the operation.
	ctx context.Context,
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	// +optional=true
	source *Directory,
	// The name of the repository. Format: owner/name.
	// +optional=true
	repo string,
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	// +optional=true
	tag string,
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	// +optional=true
	branch string,
	// Path to the workflows' directory.
	// +optional=true
	// +default=.github/workflows
	workflowsDir string,
) (*Gale, error) {
	info, err := NewRepoInfo(ctx, source, repo, tag, branch)
	if err != nil {
		return nil, err
	}

	return &Gale{
		Repo:      info,
		Workflows: info.workflows(workflowsDir),
	}, nil
}

type Gale struct {
	// The repository information
	Repo *RepoInfo

	// The workflows in the repository
	Workflows *Workflows
}

// List returns a list of workflows and their jobs with the given options.
func (g *Gale) List(ctx context.Context) (string, error) {
	workflows, err := g.Workflows.List(ctx)
	if err != nil {
		return "", err
	}

	sb := &strings.Builder{}

	var (
		indentation = "  "
		newline     = "\n"
	)

	for _, workflow := range workflows {
		sb.WriteString("- Workflow: ")
		if workflow.Name != "" {
			sb.WriteString(fmt.Sprintf("%s (path: %s)", workflow.Name, workflow.Path))
		} else {
			sb.WriteString(workflow.Path)
		}
		sb.WriteString(newline)

		sb.WriteString(indentation)
		sb.WriteString("Jobs:")
		sb.WriteString(newline)

		for _, job := range workflow.Jobs {
			sb.WriteString(indentation)
			sb.WriteString(fmt.Sprintf("  - %s", job.JobID))
			sb.WriteString(newline)
		}

		sb.WriteString("\n") // extra empty line
	}

	return sb.String(), nil
}

func (g *Gale) Run(
	// context to use for the operation
	ctx context.Context,
	// External workflow file to run.
	// +optional=true
	workflowFile *File,
	// Name of the workflow to run.
	// +optional=true
	workflow string,
	// Name of the job to run. If empty, all jobs will be run.
	// +optional=true
	job string,
	// Name of the event that triggered the workflow. e.g. push
	// +optional=true
	// +default=push
	event string,
	// File with the complete webhook event payload.
	// +optional=true
	eventFile *File,
	// Container to use for the runner(default: ghcr.io/catthehacker/ubuntu:act-latest).
	// +optional=true
	container *Container,
	// Enables debug mode.
	// +optional=true
	// +default=false
	runnerDebug bool,
	// Enables native Docker support, allowing direct execution of Docker commands in the workflow.
	// +optional=true
	// +default=true
	useNativeDocker bool,
	// Sets DOCKER_HOST to use for the native docker support.
	// +optional=true
	// +default=unix:///var/run/docker.sock
	dockerHost string,
	// Enables docker-in-dagger support to be able to run docker commands isolated from the host. Enabling DinD may lead to longer execution times.
	// +optional=true
	// +default=false
	useDind bool,
	// GitHub token to use for authentication.
	// +optional=true
	token *Secret,
) (*WorkflowRun, error) {
	if eventFile == nil {
		eventFile = dag.Directory().WithNewFile("event.json", "{}").File("event.json")
	}

	if container == nil {
		container = dag.Container().From("ghcr.io/catthehacker/ubuntu:act-latest")
	}

	if useNativeDocker && useDind {
		useNativeDocker = false
		log.Warnf("Both enableDocker and useDind are enabled. Using DinD to run docker commands.")
	}

	if useDind {
		log.Warnf("Enabling DinD may lead to longer execution times.")
	}

	planner := NewWorkflowExecutionPlanner(
		g.Repo,
		g.Workflows,
		&WorkflowRunOpts{
			WorkflowFile: workflowFile,
			Workflow:     workflow,
			Job:          job,
		},
		&RunnerOpts{
			Ctr:             container,
			Debug:           runnerDebug,
			UseNativeDocker: useNativeDocker,
			DockerHost:      dockerHost,
			UseDind:         useDind,
		},
		&EventOpts{Name: event, File: eventFile},
		&SecretOpts{Token: token},
	)

	executor, err := planner.Plan(ctx)
	if err != nil {
		return nil, err
	}

	return executor.Execute(ctx)
}

func (g *Gale) Action(
	// ID of the step. Defaults to the step index in the job.
	// +optional=true
	stepID string,
	// External workflow file to run.
	// +optional=false
	uses string,
	// Environment variables for the action. Format: name=value.
	// +optional=true
	env []string,
	// Input parameters for the action. Format: name=value.
	// +optional=true
	with []string,
) (*Actions, error) {
	actions := &Actions{
		Repo:      g.Repo,
		Workflows: g.Workflows,
	}

	return actions.Action(stepID, uses, env, with)
}
