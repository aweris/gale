package core

import (
	"fmt"

	"dagger.io/dagger"
)

// GithubRepositoryContext is a context that contains information about the repository.
type GithubRepositoryContext struct {
	Repository        string            // Repository is the combination of owner and name of the repository. e.g. octocat/hello-world
	RepositoryID      string            // RepositoryID is the id of the repository. e.g. 1296269. Note that this is different from the repository name.
	RepositoryOwner   string            // RepositoryOwner is the owner of the repository. e.g. octocat
	RepositoryOwnerID string            // RepositoryOwnerID is the id of the repository owner. e.g. 1234567. Note that this is different from the repository owner name.
	RepositoryURL     string            // RepositoryURL is the git url of the repository. e.g. git://github.com/octocat/hello-world.git.
	Workspace         string            // Workspace is the path of a directory that contains a checkout of the repository.
	Dir               *dagger.Directory // Dir is the directory where the repository is checked out
}

// NewGithubRepositoryContext creates a new GithubRepositoryContext from the given repository.
func NewGithubRepositoryContext(repo *Repository) GithubRepositoryContext {
	return GithubRepositoryContext{
		Repository:        repo.NameWithOwner,
		RepositoryID:      repo.ID,
		RepositoryOwner:   repo.Owner.Login,
		RepositoryOwnerID: repo.Owner.ID,
		RepositoryURL:     repo.URL,
		Workspace:         fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name),
		Dir:               repo.Dir,
	}
}

// Apply applies the GithubRepositoryContext to the given container.
func (c GithubRepositoryContext) Apply(container *dagger.Container) *dagger.Container {
	return container.
		WithEnvVariable("GITHUB_REPOSITORY", c.Repository).
		WithEnvVariable("GITHUB_REPOSITORY_ID", c.RepositoryID).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER", c.RepositoryOwner).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER_ID", c.RepositoryOwnerID).
		WithEnvVariable("GITHUB_REPOSITORY_URL", c.RepositoryURL).
		WithEnvVariable("GITHUB_WORKSPACE", c.Workspace).
		WithMountedDirectory(c.Workspace, c.Dir).
		WithWorkdir(c.Workspace)
}
