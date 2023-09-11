package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"

	"github.com/aweris/gale/internal/config"
)

// Repository represents a GitHub repository
type Repository struct {
	ID               string
	Name             string
	NameWithOwner    string
	URL              string
	Owner            RepositoryOwner
	DefaultBranchRef RepositoryBranchRef
	GitRef           *RepositoryGitRef
}

// RepositoryOwner represents a GitHub repository owner
type RepositoryOwner struct {
	ID    string
	Login string
}

// RepositoryBranchRef represents a GitHub repository branch ref
type RepositoryBranchRef struct {
	Name string
}

// GetRepositoryOpts represents the options for refs used to get a repository. Only one of the
// options can be set.
//
// If none of the options are set, the default branch will be used. Default branch or remote repositories is configured
// in the GitHub repository settings. For local repositories, it's the branch that is currently checked out.
//
// If multiple options are set, the precedence is as follows: tag, branch.
type GetRepositoryOpts struct {
	Branch string
	Tag    string
}

// GetCurrentRepository returns current repository information. This is a wrapper around GetRepository with empty name.
func GetCurrentRepository(opts ...GetRepositoryOpts) (*Repository, error) {
	return GetRepository("", opts...)
}

// GetRepository returns repository information. If name is empty, the current repository will be used.
func GetRepository(name string, opts ...GetRepositoryOpts) (*Repository, error) {
	var repo Repository

	stdout, stderr, err := gh.Exec("repo", "view", name, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef")
	if err != nil {
		return nil, fmt.Errorf("failed to get current repository: %w stderr: %s", err, stderr.String())
	}

	err = json.Unmarshal(stdout.Bytes(), &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal current repository: %s err: %w", stdout.String(), err)
	}

	opt := GetRepositoryOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// load repo tree based on the options precedence
	switch {
	case opt.Tag != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeTag, opt.Tag)
		if err != nil {
			return nil, err
		}
	case opt.Branch != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeBranch, opt.Branch)
		if err != nil {
			return nil, err
		}
	case name != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeBranch, repo.DefaultBranchRef.Name)
		if err != nil {
			return nil, err
		}
	default:
		// TODO: current directory could be a subdirectory of the repository. Should we handle this case?
		repo.GitRef, err = GetRepositoryRefFromDir(context.Background(), config.Client().Host().Directory("."))
		if err != nil {
			return nil, err
		}
	}

	return &repo, nil
}
