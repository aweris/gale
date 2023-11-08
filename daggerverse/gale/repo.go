package main

import (
	"context"
	"path/filepath"
	"strings"
)

// getRepoInfo returns the repository info from the given options.
func getRepoInfo(source Optional[*Directory], repo, tag, branch Optional[string]) *RepoInfo {
	// convert workflows list options to repo source options
	opts := RepoInfoOpts{
		Source: source.GetOr(nil),
		Repo:   repo.GetOr(""),
		Tag:    tag.GetOr(""),
		Branch: branch.GetOr(""),
	}

	// get the repository source working directory from the options -- default value handled by the repo module, so we
	// don't need to handle it here.
	return dag.Repo().Info(opts)
}

// getWorkflowsDir returns the workflows directory from the given options.
func getWorkflowsDir(source Optional[*Directory], repo, tag, branch, workflowsDir Optional[string]) *Directory {
	info := getRepoInfo(source, repo, tag, branch)

	// get the repository source working directory from the options -- default value handled by the repo module, so we
	// don't need to handle it here.
	return info.WorkflowsDir(RepoInfoWorkflowsDirOpts{WorkflowsDir: workflowsDir.GetOr("")})
}

// WorkflowWalkFunc is the type of the function called for each workflow file visited by walkWorkflowDir.
type WorkflowWalkFunc func(ctx context.Context, path string, file *File) error

// walkWorkflowDir walks the workflows directory and calls the given function for each workflow file.
func walkWorkflowDir(ctx context.Context, path Optional[string], dir *Directory, fn WorkflowWalkFunc) error {
	entries, err := dir.Entries(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(path.GetOr(".github/workflows"), entry)

			if err := fn(ctx, path, file); err != nil {
				return err
			}
		}
	}

	return nil
}
