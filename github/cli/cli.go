package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
)

// Repository represents a GitHub repository
type Repository struct {
	ID               string
	Name             string
	NameWithOwner    string
	URL              string
	Owner            RepositoryOwner
	DefaultBranchRef RepositoryBranchRef
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

// CurrentRepository returns current repository information. This is a wrapper around
// gh repo view --json id,name,owner,nameWithOwner,url,defaultBranchRef
func CurrentRepository(ctx context.Context) (*Repository, error) {
	stdout, stderr, err := gh.ExecContext(
		ctx, "repo", "view", "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get current repository: %w stderr: %s", err, stderr.String())
	}

	var repo Repository

	err = json.Unmarshal(stdout.Bytes(), &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal current repository: %s err: %w", stdout.String(), err)
	}

	return &repo, nil
}
