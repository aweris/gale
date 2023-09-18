package context

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/aweris/gale/ghx/core"
)

// errRefFound is a sentinel error used to stop iteration early.
var errRefFound = errors.New("stop")

func (r *RepositoryContext) LoadFromDirectory(path string) error {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get current reference (branch, tag, or commit)
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get head reference: %w", err)
	}

	if head.Hash().IsZero() {
		return fmt.Errorf("failed to get head reference: %w", plumbing.ErrReferenceNotFound)
	}

	iter, err := repo.References()
	if err != nil {
		return fmt.Errorf("failed to get references: %w", err)
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
		return err
	}

	switch {
	case found.Name().IsBranch():
		r.RefType = core.RefTypeBranch
		r.Ref = found.Name().String()
		r.RefName = found.Name().Short()
		r.IsRemote = found.Name().IsRemote()
	case found.Name().IsTag():
		r.RefType = core.RefTypeTag
		r.Ref = found.Name().String()
		r.RefName = found.Name().Short()
		r.IsRemote = found.Name().IsRemote()
	default:
		return fmt.Errorf("unsupported ref: %s", found.Name().String())
	}

	return nil
}
