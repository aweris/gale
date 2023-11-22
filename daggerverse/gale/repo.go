package main

import (
	"context"
	"path/filepath"
	"strings"
)

// RepoInfoOpts is a set of options for getting repository information.
func toRepoInfoOpts(source Optional[*Directory], repo, branch, tag Optional[string]) RepoInfoOpts {
	return RepoInfoOpts{
		Source: source.GetOr(nil),
		Repo:   repo.GetOr(""),
		Branch: branch.GetOr(""),
		Tag:    tag.GetOr(""),
	}
}

// getWorkflowsDir returns the workflows directory from the given options.
func getWorkflowsDir(source Optional[*Directory], repo, tag, branch, workflowsDir Optional[string]) *Directory {
	return dag.Repo().
		Info(toRepoInfoOpts(source, repo, branch, tag)).
		Source().
		Directory(workflowsDir.GetOr(".github/workflows"))
}

// workflowWalkFunc is the type of the function called for each workflow file visited by walkWorkflowDir.
type workflowWalkFunc func(ctx context.Context, path string, file *File) (bool, error)

// walkWorkflowDir walks the workflows directory and calls the given function for each workflow file. If walk function
// returns false, the walk stops.
func walkWorkflowDir(ctx context.Context, path Optional[string], dir *Directory, fn workflowWalkFunc) error {
	entries, err := dir.Entries(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)
			path := filepath.Join(path.GetOr(".github/workflows"), entry)

			walk, err := fn(ctx, path, file)
			if err != nil {
				return err
			}

			if !walk {
				return nil
			}
		}
	}

	return nil
}
