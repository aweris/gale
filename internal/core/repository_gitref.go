package core

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
	"github.com/aweris/gale/internal/config"
)

// RefType represents the type of a ref. It can be either a branch or a tag.
type RefType string

const (
	RefTypeBranch RefType = "branch"
	RefTypeTag    RefType = "tag"
)

// RepositoryGitRef represents a Git ref (branch or tag) in a repository
type RepositoryGitRef struct {
	Ref     string
	RefName string
	RefType RefType
	Dir     *dagger.Directory
}

// GetRepositoryGitRef returns a Git ref (branch or tag) in a repository. If name is empty, the current repository will be used.
func GetRepositoryGitRef(_ context.Context, url string, refType RefType, refName string) (*RepositoryGitRef, error) {
	var (
		ref string
		dir *dagger.Directory
	)

	git := config.Client().Git(url, dagger.GitOpts{KeepGitDir: true})

	switch refType {
	case RefTypeBranch:
		dir = git.Branch(refName).Tree()
		ref = fmt.Sprintf("refs/heads/%s", refName)
	case RefTypeTag:
		dir = git.Tag(refName).Tree()
		ref = fmt.Sprintf("refs/tags/%s", refName)
	default:
		return nil, fmt.Errorf("invalid ref type: %s", refType)
	}

	return &RepositoryGitRef{Ref: ref, RefName: refName, RefType: refType, Dir: dir}, nil
}

// GetRepositoryRefFromDir returns a Git ref (branch or tag) from given directory. If dir is empty or not git repository, it will return an error.
func GetRepositoryRefFromDir(ctx context.Context, dir *dagger.Directory) (*RepositoryGitRef, error) {
	var (
		ref     string
		refType RefType
		refName string
	)

	out, err := config.Client().
		Container().
		From("alpine/git").
		WithMountedDirectory("/src", dir).WithWorkdir("/src").
		WithExec([]string{"rev-parse", "--symbolic-full-name", "HEAD"}).
		Stdout(ctx)
	if err != nil {
		return nil, err
	}

	ref = strings.TrimSpace(out)

	switch {
	case strings.HasPrefix(ref, "refs/heads/"):
		refType = RefTypeBranch
		refName = strings.TrimPrefix(ref, "refs/heads/")
	case strings.HasPrefix(ref, "refs/tags/"):
		refType = RefTypeTag
		refName = strings.TrimPrefix(ref, "refs/tags/")
	default:
		return nil, fmt.Errorf("invalid ref type: %s", refType)
	}

	return &RepositoryGitRef{Ref: ref, RefName: refName, RefType: refType, Dir: dir}, nil
}
