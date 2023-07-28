package core

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
)

type RunnerContext struct {
	Name      string `json:"name"`       // Name is the name of the runner.
	OS        string `json:"os"`         // OS is the operating system of the runner.
	Arch      string `json:"arch"`       // Arch is the architecture of the runner.
	Temp      string `json:"temp"`       // Temp is the path to a temporary directory on the runner.
	ToolCache string `json:"tool_cache"` // ToolCache is the path to the directory containing preinstalled tools for GitHub-hosted runners.
	Debug     string `json:"debug"`      // Debug is set only if debug logging is enabled, and always has the value of 1.
}

// NewRunnerContext creates a new RunnerContext from the given runner.
func NewRunnerContext() RunnerContext {
	return RunnerContext{
		Name:      "Gale Agent",
		OS:        "linux",
		Arch:      "x64",
		Temp:      "/home/runner/_temp",
		ToolCache: "/opt/hostedtoolcache",
		Debug:     "0", // TODO: This should be configurable. Read from config.Debug()
	}
}

// Apply applies the RunnerContext to the given container.
func (c RunnerContext) Apply(container *dagger.Container) *dagger.Container {
	return container.
		WithEnvVariable("RUNNER_NAME", c.Name).
		WithEnvVariable("RUNNER_TEMP", c.Temp).
		WithEnvVariable("RUNNER_OS", c.OS).
		WithEnvVariable("RUNNER_ARCH", c.Arch).
		WithEnvVariable("RUNNER_TOOL_CACHE", c.ToolCache).
		WithEnvVariable("RUNNER_DEBUG", c.Debug)
}

// GithubRepositoryContext is a context that contains information about the repository.
type GithubRepositoryContext struct {
	Repository        string            `json:"repository"`          // Repository is the combination of owner and name of the repository. e.g. octocat/hello-world
	RepositoryID      string            `json:"repository_id"`       // RepositoryID is the id of the repository. e.g. 1296269. Note that this is different from the repository name.
	RepositoryOwner   string            `json:"repository_owner"`    // RepositoryOwner is the owner of the repository. e.g. octocat
	RepositoryOwnerID string            `json:"repository_owner_id"` // RepositoryOwnerID is the id of the repository owner. e.g. 1234567. Note that this is different from the repository owner name.
	RepositoryURL     string            `json:"repository_url"`      // RepositoryURL is the git url of the repository. e.g. git://github.com/octocat/hello-world.git.
	Workspace         string            `json:"workspace"`           // Workspace is the path of a directory that contains a checkout of the repository.
	Dir               *dagger.Directory `json:"-"`                   // Dir is the directory where the repository is checked out
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

// GithubSecretsContext is a context that contains information about the secrets.
type GithubSecretsContext struct {
	Token string `json:"token"` // Token is the GitHub token to use for authentication.
}

// NewGithubSecretsContext creates a new GithubSecretsContext from the given token.
func NewGithubSecretsContext(token string) GithubSecretsContext {
	return GithubSecretsContext{
		Token: token,
	}
}

// Apply applies the GithubSecretsContext to the given container.
func (c GithubSecretsContext) Apply(container *dagger.Container) *dagger.Container {
	return container.WithSecretVariable("GITHUB_TOKEN", config.Client().SetSecret("GITHUB_TOKEN", c.Token))
}

// GithubURLContext is a context that contains URLs for the Github server and API.
type GithubURLContext struct {
	ApiURL     string `json:"api_url"`     // ApiURL is the URL of the Github API. e.g. https://api.github.com
	GraphqlURL string `json:"graphql_url"` // GraphqlURL is the URL of the Github GraphQL API. e.g. https://api.github.com/graphql
	ServerURL  string `json:"server_url"`  // ServerURL is the URL of the Github server. e.g. https://github.com
}

// NewGithubURLContext creates a new GithubURLContext from the given urls.
func NewGithubURLContext() GithubURLContext {
	return GithubURLContext{
		ApiURL:     "https://api.github.com",
		GraphqlURL: "https://api.github.com/graphql",
		ServerURL:  "https://github.com",
	}
}

// Apply applies the GithubURLContext to the given container.
func (c GithubURLContext) Apply(container *dagger.Container) *dagger.Container {
	return container.
		WithEnvVariable("GITHUB_API_URL", c.ApiURL).
		WithEnvVariable("GITHUB_GRAPHQL_URL", c.GraphqlURL).
		WithEnvVariable("GITHUB_SERVER_URL", c.ServerURL)
}

// GithubFilesContext is a context that contains paths for files and directories useful to Github Actions.
//
// This context changes per step. It'll be used on ghx level to create a temporary file that sets environment variables
// from workflow commands. It won't be used or applied to the container.
type GithubFilesContext struct {
	Env  string `json:"env"`  // Env is the path to a temporary file that sets environment variables from workflow commands.
	Path string `json:"path"` // Path is the path to a temporary file that sets the system PATH variable from workflow commands.
}

// StepContext is a context that contains information about the step.
//
// This context created per step execution. It won't be used or applied to the container level.
type StepContext struct {
	Conclusion Conclusion        `json:"conclusion"` // Conclusion is the result of a completed step after continue-on-error is applied
	Outcome    Conclusion        `json:"outcome"`    // Outcome is  the result of a completed step before continue-on-error is applied
	Outputs    map[string]string `json:"outputs"`    // Outputs is a map of output name to output value
	State      map[string]string `json:"-"`          // State is a map of step state variables. This is not available to expressions so that's why json tag is set to "-" to ignore it.
	Summary    string            `json:"-"`          // Summary is the summary of the step. This is not available to expressions so that's why json tag is set to "-" to ignore it.
}
