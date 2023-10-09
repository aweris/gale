package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/cli/go-gh/v2/pkg/repository"
)

// Repository represents a GitHub repository
type Repository struct {
	ID            int             `json:"id" env:"GALE_REPO_ID" container_env:"true"`
	Name          string          `json:"name" env:"GALE_REPO_NAME" container_env:"true"`
	FullName      string          `json:"full_name" env:"GALE_REPO_NAME_WITH_OWNER" container_env:"true"`
	CloneURL      string          `json:"clone_url" env:"GALE_REPO_URL" container_env:"true"`
	Owner         RepositoryOwner `json:"owner" container_env:"true"`
	DefaultBranch string          `json:"default_branch" env:"GALE_REPO_BRANCH_NAME" container_env:"true"`
}

// RepositoryOwner represents a GitHub repository owner
type RepositoryOwner struct {
	ID    int    `env:"GALE_REPO_OWNER_ID" container_env:"true"`
	Login string `env:"GALE_REPO_OWNER_LOGIN" container_env:"true"`
}

// GetRepository returns repository information. If name is empty, the current repository will be used.
func GetRepository(name string) (Repository, error) {
	var repo Repository

	if name == "" {
		current, err := repository.Current()
		if err != nil {
			return repo, fmt.Errorf("failed to get current repository: %w", err)
		}

		name = fmt.Sprintf("%s/%s", current.Owner, current.Name)

		// normalize the name
		name = strings.TrimSpace(name)
		name = strings.Trim(name, "/")
	}

	opts := api.ClientOptions{}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		opts.AuthToken = token
	}

	client, err := api.NewRESTClient(opts)
	if err != nil {
		return repo, fmt.Errorf("failed to create github client: %w", err)
	}

	err = client.Get(fmt.Sprintf("repos/%s", name), &repo)
	if err != nil {
		return repo, fmt.Errorf("failed to get repository: %w", err)
	}

	return repo, nil
}
