package main

import (
	"context"
	"fmt"
	"strings"
)

func New(
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// Path to the workflows' directory. (default: .github/workflows)
	workflowsDir Optional[string],
) (*Gale, error) {
	gale := &Gale{
		Source:       withEmptyValue(source),
		Repo:         withEmptyValue(repo),
		Tag:          withEmptyValue(tag),
		Branch:       withEmptyValue(branch),
		WorkflowsDir: workflowsDir.GetOr(".github/workflows"),
	}

	// validate options

	if gale.Source == nil && gale.Repo == "" {
		return nil, fmt.Errorf("either a repo or a source directory must be provided")
	}

	if gale.Repo != "" && gale.Branch == "" && gale.Tag == "" {
		return nil, fmt.Errorf("when repo is provided, either a branch or a tag must be provided")
	}

	return gale, nil
}

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct {
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	Source *Directory

	// The name of the repository. Format: owner/name.
	Repo string

	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Tag string

	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Branch string

	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	WorkflowsDir string
}

// List returns a list of workflows and their jobs with the given options.
func (g *Gale) List(ctx context.Context) (string, error) {
	// load repository information
	info, err := internal.RepoInfo(ctx, g.Source, g.Repo, g.Tag, g.Branch)
	if err != nil {
		return "", err
	}

	workflows, err := info.workflows().List(ctx, g.WorkflowsDir)
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
			sb.WriteString(fmt.Sprintf("%s", workflow.Path))
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
	workflowFile Optional[*File],
	// Name of the workflow to run.
	workflow Optional[string],
	// Name of the job to run. If empty, all jobs will be run.
	job Optional[string],
	// Name of the event that triggered the workflow. e.g. push
	event Optional[string],
	// File with the complete webhook event payload.
	eventFile Optional[*File],
	// Container to use for the runner(default: ghcr.io/catthehacker/ubuntu:act-latest).
	container Optional[*Container],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token to use for authentication.
	token Optional[*Secret],
) (*WorkflowRun, error) {
	plan, err := internal.WorkflowExecutionPlan(
		ctx,
		g.Source,
		g.Repo,
		g.Tag,
		g.Branch,
		g.WorkflowsDir,
		workflowFile,
		workflow,
		job,
		container,
		event,
		eventFile,
		runnerDebug,
		token,
	)
	if err != nil {
		return nil, err
	}

	return plan.Execute(ctx)
}
