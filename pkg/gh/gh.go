package gh

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"

	"github.com/aweris/gale/pkg/model"
)

// CurrentRepository returns current repository information. This is a wrapper around
// gh repository view --json id,name,owner,nameWithOwner,url,defaultBranchRef
func CurrentRepository() (*model.Repository, error) {
	var repo model.Repository

	stdout, stderr, err := gh.Exec("repo", "view", "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef")
	if err != nil {
		return nil, fmt.Errorf("failed to get current repository: %w stderr: %s", err, stderr.String())
	}

	err = json.Unmarshal(stdout.Bytes(), &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal current repository: %s err: %w", stdout.String(), err)
	}

	return &repo, nil
}

// CurrentUser returns current user information
func CurrentUser() (*model.User, error) {
	stdout, stderr, err := gh.Exec("api", "user")
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w stderr: %s", err, stderr.String())
	}

	var user model.User

	err = json.Unmarshal(stdout.Bytes(), &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal current user: %s err: %w", stdout.String(), err)
	}

	return &user, nil
}

// GetToken returns the auth token gh is configured to use
func GetToken() (string, error) {
	stdout, stderr, err := gh.Exec("auth", "token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w stderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
