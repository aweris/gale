package main

import (
	"context"
	"encoding/json"
	"fmt"
)

// unmarshalContentsToJSON unmarshal the contents of the file as JSON into the given value.
func unmarshalContentsToJSON(ctx context.Context, f *File, v interface{}) error {
	stdout, err := f.Contents(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get file contents", err)
	}

	err = json.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal file contents", err)
	}

	return nil
}

// RepoInfoOpts is a set of options for getting repository information.
func toRepoInfoOpts(source Optional[*Directory], repo, branch, tag Optional[string]) RepoInfoOpts {
	return RepoInfoOpts{
		Source: source.GetOr(nil),
		Repo:   repo.GetOr(""),
		Branch: branch.GetOr(""),
		Tag:    tag.GetOr(""),
	}
}
