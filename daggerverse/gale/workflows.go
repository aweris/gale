package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Workflows struct{}

func (w *Workflows) List(
	ctx context.Context,
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
) (string, error) {
	// convert workflows list options to repo source options
	opts := RepoInfoOpts{
		Source: source.GetOr(nil),
		Repo:   repo.GetOr(""),
		Tag:    tag.GetOr(""),
		Branch: branch.GetOr(""),
	}

	// get the repository source working directory from the options
	dir := dag.Repo().
		Info(opts).
		Source().
		Directory(workflowsDir.GetOr(".github/workflows"))

	// list all entries in the workflows directory
	entries, err := dir.Entries(ctx)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(workflowsDir.GetOr(".github/workflows"), entry)

			// dagger do not support maps yet, so we're defining anonymous struct to unmarshal the yaml file to avoid
			// hit this limitation.

			var workflow struct {
				Name string                 `yaml:"name"`
				Jobs map[string]interface{} `yaml:"jobs"`
			}

			if err := unmarshalContentsToYAML(ctx, file, &workflow); err != nil {
				return "", err
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
		}
	}

	return sb.String(), nil
}

func (w *Workflows) Run(
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
			RunnerImage:  runnerImage.GetOr("ghcr.io/catthehacker/ubuntu:act-latest"),
			RunnerDebug:  runnerDebug.GetOr(false),
			Token:        token.GetOr(nil),
		},
	}
}
