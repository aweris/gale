package core

import (
	"encoding/json"
	"fmt"

	"github.com/cli/go-gh/v2"
)

// Repository represents a GitHub repository
type Repository struct {
	ID               string              `env:"GALE_REPO_ID" container_env:"true"`
	Name             string              `env:"GALE_REPO_NAME" container_env:"true"`
	NameWithOwner    string              `env:"GALE_REPO_NAME_WITH_OWNER" container_env:"true"`
	URL              string              `env:"GALE_REPO_URL" container_env:"true"`
	Owner            RepositoryOwner     `container_env:"true"`
	DefaultBranchRef RepositoryBranchRef `container_env:"true"`
}

// RepositoryOwner represents a GitHub repository owner
type RepositoryOwner struct {
	ID    string `env:"GALE_REPO_OWNER_ID" container_env:"true"`
	Login string `env:"GALE_REPO_OWNER_LOGIN" container_env:"true"`
}

// RepositoryBranchRef represents a GitHub repository branch ref
type RepositoryBranchRef struct {
	Name string `env:"GALE_REPO_BRANCH_NAME" container_env:"true"`
}

// GetRepository returns repository information. If name is empty, the current repository will be used.
func GetRepository(name string) (Repository, error) {
	var repo Repository

	stdout, stderr, err := gh.Exec("repo", "view", name, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef")
	if err != nil {
		return repo, fmt.Errorf("failed to get current repository: %w stderr: %s", err, stderr.String())
	}

	err = json.Unmarshal(stdout.Bytes(), &repo)
	if err != nil {
		return repo, fmt.Errorf("failed to unmarshal current repository: %s err: %w", stdout.String(), err)
	}

	return repo, nil
}
