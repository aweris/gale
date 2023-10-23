package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

type Workflows struct{}

func (w *Workflows) List(ctx context.Context, repoOpts RepoOpts, dirOpts WorkflowsDirOpts) (string, error) {
	dir := dag.Repo().Source((RepoSourceOpts)(repoOpts)).Directory(dirOpts.WorkflowsDir)

	entries, err := dir.Entries(ctx)
	if err != nil {
		return "", err
	}

	sb := strings.Builder{}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(dirOpts.WorkflowsDir, entry)

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

func (w *Workflows) Run(repoOpts RepoOpts, pathOpts WorkflowsDirOpts, runOpts WorkflowsRunOpts) *WorkflowRun {
	return &WorkflowRun{
		Config: WorkflowRunConfig{
			Source:       repoOpts.Source,
			Repo:         repoOpts.Repo,
			Branch:       repoOpts.Branch,
			Tag:          repoOpts.Tag,
			WorkflowsDir: pathOpts.WorkflowsDir,
			Workflow:     runOpts.Workflow,
			Job:          runOpts.Job,
			Event:        runOpts.Event,
			EventFile:    runOpts.EventFile,
			RunnerImage:  runOpts.RunnerImage,
			RunnerDebug:  runOpts.RunnerDebug,
			Token:        runOpts.Token,
		},
	}
}
