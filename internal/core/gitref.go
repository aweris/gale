package core

import (
	"github.com/cli/go-gh/v2"
)

// RefType represents the type of a ref. It can be either a branch or a tag.
type RefType string

const (
	RefTypeBranch  RefType = "branch"
	RefTypeTag     RefType = "tag"
	RefTypeCommit  RefType = "commit"
	RefTypeUnknown RefType = "unknown"
)

// DetermineRefTypeFromRepo determines the type of ref from given repository and ref. The ref can be either a branch or a tag.
// If the ref is not a branch, a tag or a commit, it will return RefTypeUnknown.
//
// The method will use GitHub API to determine the type of ref. If the ref does not exist on remote, it will
// return RefTypeUnknown.
func DetermineRefTypeFromRepo(repo, ref string) RefType {
	// make api call to github to determine the type of ref. we don't interest in the response, we just want to know
	// if the ref exists or not. If the ref does not exist, the api call will return an error.
	_, _, err := gh.Exec("api", "repos/"+repo+"/git/ref/heads/"+ref)
	if err == nil {
		return RefTypeBranch
	}

	_, _, err = gh.Exec("api", "repos/"+repo+"/git/ref/tags/"+ref)
	if err == nil {
		return RefTypeTag
	}

	_, _, err = gh.Exec("api", "repos/"+repo+"/git/commits/"+ref)
	if err == nil {
		return RefTypeCommit
	}

	return RefTypeUnknown
}
