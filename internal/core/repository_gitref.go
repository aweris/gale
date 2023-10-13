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
func GetRepositoryRefFromDir(path string) (RepositoryGitRef, error) {
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
