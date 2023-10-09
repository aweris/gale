package core

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
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
	client, err := api.DefaultRESTClient()
	if err != nil {
		return RefTypeUnknown
	}

	var dummy interface{}

	err = client.Get(fmt.Sprintf("repos/%s/git/ref/heads/%s", repo, ref), &dummy)
	if err == nil {
		return RefTypeBranch
	}

	err = client.Get(fmt.Sprintf("repos/%s/git/ref/tags/%s", repo, ref), &dummy)
	if err == nil {
		return RefTypeTag
	}

	err = client.Get(fmt.Sprintf("repos/%s/git/commits/%s", repo, ref), &dummy)
	if err == nil {
		return RefTypeCommit
	}

	return RefTypeUnknown
}
