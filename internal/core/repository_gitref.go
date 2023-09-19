package core

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// errRefFound is a sentinel error used to stop iteration early.
var errRefFound = errors.New("stop")

// RepositoryGitRef represents a Git ref (branch or tag) in a repository
type RepositoryGitRef struct {
	Ref      string  `json:"ref" env:"GALE_REPO_REF" container_env:"true"`
	RefName  string  `json:"ref_name" env:"GALE_REPO_REF_NAME" container_env:"true"`
	RefType  RefType `json:"ref_type" env:"GALE_REPO_REF_TYPE" container_env:"true"`
	SHA      string  `json:"sha" env:"GALE_REPO_SHA" container_env:"true"`
	ShortSHA string  `json:"short_sha" env:"GALE_REPO_SHORT_SHA" container_env:"true"`
	IsRemote bool    `json:"is_remote" env:"GALE_REPO_IS_REMOTE" container_env:"true"`
}

// GetRepositoryRefFromDir returns a Git ref (branch or tag) from given directory. If dir is empty or not git repository, it will return an error.
func GetRepositoryRefFromDir(path, tag, branch string) (RepositoryGitRef, error) {
	var ref RepositoryGitRef

	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return ref, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current reference (branch, tag, or commit)
	head, err := repo.Head()
	if err != nil {
		return ref, fmt.Errorf("failed to get head reference: %w", err)
	}

	if head.Hash().IsZero() {
		return ref, fmt.Errorf("failed to get head reference: %w", plumbing.ErrReferenceNotFound)
	}

	iter, err := repo.References()
	if err != nil {
		return ref, fmt.Errorf("failed to get references: %w", err)
	}

	var found *plumbing.Reference

	// Iterate through all references to find the matching one
	err = iter.ForEach(func(ref *plumbing.Reference) error {
		if ref.Hash() == head.Hash() {
			found = ref

			return errRefFound
		}

		return nil
	})

	if err != nil && !errors.Is(err, errRefFound) {
		return ref, err
	}

	var (
		refType  RefType
		refFull  string
		refName  string
		isRemote bool
	)

	switch {
	case found.Name().IsBranch():
		refType = RefTypeBranch
		refFull = found.Name().String()
		refName = found.Name().Short()
		isRemote = found.Name().IsRemote()
	case found.Name().IsTag():
		refType = RefTypeTag
		refFull = found.Name().String()
		refName = found.Name().Short()
		isRemote = found.Name().IsRemote()
	case found.Name() == "HEAD" && tag != "": // we're in detached head state and tag is provided, so we're using the tag
		refType = RefTypeTag
		refFull = fmt.Sprintf("refs/tags/%s", tag)
		refName = tag
		isRemote = true
	case found.Name() == "HEAD" && branch != "": // we're in detached head state and branch is provided, so we're using the branch
		refType = RefTypeBranch
		refFull = fmt.Sprintf("refs/heads/%s", branch)
		refName = branch
		isRemote = true
	default:
		return ref, fmt.Errorf("unsupported ref: %s", found.Name().String())
	}

	return RepositoryGitRef{
		Ref:      refFull,
		RefName:  refName,
		RefType:  refType,
		SHA:      found.Hash().String(),
		ShortSHA: found.Hash().String()[:7],
		IsRemote: isRemote,
	}, nil
}
