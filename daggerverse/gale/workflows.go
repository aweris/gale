package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Workflows struct{}

// WorkflowsRepoOpts represents the options for getting repository information.
//
// This is copy of RepoOpts from daggerverse/gale/repo.go to be able to expose options with gale module and pass them to
// the repo module just type casting.
type WorkflowsRepoOpts struct {
	Source *Directory `doc:"The directory containing the repository source. If source is provided, rest of the options are ignored."`
	Repo   string     `doc:"The name of the repository. Format: owner/name."`
	Branch string     `doc:"Branch name to checkout. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
	Tag    string     `doc:"Tag name to checkout. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
}

// WorkflowsDirOpts represents the options for getting workflow information.
type WorkflowsDirOpts struct {
	WorkflowsDir string `doc:"The relative path to the workflow directory." default:".github/workflows"`
}

func (w *Workflows) List(ctx context.Context, repoOpts WorkflowsRepoOpts, pathOpts WorkflowsDirOpts) (string, error) {
	dir := dag.Repo().Source((RepoSourceOpts)(repoOpts)).Directory(pathOpts.WorkflowsDir)

	entries, err := dir.Entries(ctx)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(pathOpts.WorkflowsDir, entry)

			// dagger do not support maps yet, so we're defining anonymous struct to unmarshal the yaml file to avoid
			// hit this limitation.

			var workflow struct {
				Name string                 `yaml:"name"`
				Jobs map[string]interface{} `yaml:"jobs"`
			}

			if err := file.unmarshalContentsToYAML(ctx, &workflow); err != nil {
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

func (w *Workflows) Run(repoOpts WorkflowsRepoOpts, pathOpts WorkflowsDirOpts, runOpts WorkflowsRunOpts) *WorkflowRun {
	return &WorkflowRun{
		Config: &WorkflowRunConfig{
			WorkflowsRepoOpts: &repoOpts,
			WorkflowsDirOpts:  &pathOpts,
			WorkflowsRunOpts:  &runOpts,
		},
	}
}
