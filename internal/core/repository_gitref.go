package core

import (
	"context"
	"fmt"
	"strings"

	"dagger.io/dagger"
)

// RepositoryGitRef represents a Git ref (branch or tag) in a repository
type RepositoryGitRef struct {
	Ref     string
	RefName string
	RefType RefType
	SHA     string
	Dir     *dagger.Directory
}

// GetRepositoryGitRef returns a Git ref (branch or tag) in a repository. If name is empty, the current repository will be used.
func GetRepositoryGitRef(ctx context.Context, client *dagger.Client, url string, refType RefType, refName string) (*RepositoryGitRef, error) {
	var (
		ref string
		dir *dagger.Directory
	)

	git := client.Git(url, dagger.GitOpts{KeepGitDir: true})

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

	sha, err := getRepoSHA(ctx, client, dir)
	if err != nil {
		return nil, err
	}

	return &RepositoryGitRef{Ref: ref, RefName: refName, RefType: refType, SHA: sha, Dir: dir}, nil
}

// GetRepositoryRefFromDir returns a Git ref (branch or tag) from given directory. If dir is empty or not git repository, it will return an error.
func GetRepositoryRefFromDir(ctx context.Context, client *dagger.Client, dir *dagger.Directory) (*RepositoryGitRef, error) {
	var (
		ref     string
		refType RefType
		refName string
		sha     string
	)

	out, err := client.
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

	sha, err = getRepoSHA(ctx, client, dir)
	if err != nil {
		return nil, err
	}

	return &RepositoryGitRef{Ref: ref, RefName: refName, RefType: refType, SHA: sha, Dir: dir}, nil
}

func getRepoSHA(ctx context.Context, client *dagger.Client, dir *dagger.Directory) (string, error) {
	out, err := client.
		Container().
		From("alpine/git").
		WithMountedDirectory("/src", dir).WithWorkdir("/src").
		WithExec([]string{"rev-parse", "HEAD"}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}
