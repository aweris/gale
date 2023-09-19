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
	Ref      string
	RefName  string
	RefType  RefType
	SHA      string
	ShortSHA string
	IsRemote bool
}

// GetRepositoryRefFromDir returns a Git ref (branch or tag) from given directory. If dir is empty or not git repository, it will return an error.
func GetRepositoryRefFromDir(path string) (*RepositoryGitRef, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current reference (branch, tag, or commit)
	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get head reference: %w", err)
	}

	if head.Hash().IsZero() {
		return nil, fmt.Errorf("failed to get head reference: %w", plumbing.ErrReferenceNotFound)
	}

	iter, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("failed to get references: %w", err)
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
		return nil, err
	}

	var refType RefType

	switch {
	case found.Name().IsBranch():
		refType = RefTypeBranch
	case found.Name().IsTag():
		refType = RefTypeTag
	default:
		return nil, fmt.Errorf("unsupported ref type: %s", found.Name().String())
	}

	return &RepositoryGitRef{
		Ref:      found.Name().String(),
		RefName:  found.Name().Short(),
		RefType:  refType,
		SHA:      found.Hash().String(),
		ShortSHA: found.Hash().String()[:7],
		IsRemote: found.Name().IsRemote(),
	}, nil
}
