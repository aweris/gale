package main

import (
	"context"
	"fmt"
	"strings"
)

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct{}

// List returns a list of workflows and their jobs with the given options.
func (g *Gale) List(
	// context to use for the operation
	ctx context.Context,
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
) (string, error) {
	workflows := getWorkflowsDir(source, repo, tag, branch, workflowsDir)

	sb := strings.Builder{}

	err := walkWorkflowDir(ctx, workflowsDir, workflows,
		func(ctx context.Context, path string, file *File) error {
			// dagger do not support maps yet, so we're defining anonymous struct to unmarshal the yaml file to avoid
			// hit this limitation.
			var workflow struct {
				Name string                 `yaml:"name"`
				Jobs map[string]interface{} `yaml:"jobs"`
			}

			if err := unmarshalContentsToYAML(ctx, file, &workflow); err != nil {
				return err
			}

			sb.WriteString("Workflow: ")
			if workflow.Name != "" {
				sb.WriteString(fmt.Sprintf("%s (path: %s)\n", workflow.Name, path))
			} else {
				sb.WriteString(fmt.Sprintf("%s\n", path))
			}

			sb.WriteString("Jobs:\n")

			for job := range workflow.Jobs {
				sb.WriteString(fmt.Sprintf(" - %s\n", job))
			}

			sb.WriteString("\n") // extra empty line

			return nil
		},
	)

	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

// Run runs the workflow with the given options.
func (g *Gale) Run(
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// Path to the workflows directory. (default: .github/workflows)
	workflowsDir Optional[string],
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
	// Image to use for the runner.
	runnerImage Optional[string],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token to use for authentication.
	token Optional[*Secret],
) *WorkflowRun {
	return &WorkflowRun{
		Config: WorkflowRunConfig{
			Base:         dag.Container().From(runnerImage.GetOr("ghcr.io/catthehacker/ubuntu:act-latest")),
			Source:       source.GetOr(nil),
			Repo:         repo.GetOr(""),
			Branch:       branch.GetOr(""),
			Tag:          tag.GetOr(""),
			WorkflowsDir: workflowsDir.GetOr(".github/workflows"),
			WorkflowFile: workflowFile.GetOr(nil),
			Workflow:     workflow.GetOr(""),
			Job:          job.GetOr(""),
			Event:        event.GetOr("push"),
			EventFile:    eventFile.GetOr(nil),
			RunnerDebug:  runnerDebug.GetOr(false),
			Token:        token.GetOr(nil),
		},
	}
}
