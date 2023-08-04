package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

var _ helpers.WithContainerFuncHook = new(RunnerContext)

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
		ToolCache: "/home/runner/hostedtoolcache", // /opt/hostedtoolcache is used by our base runner image and if we mount it we'll lose the tools installed by the base image.
		Debug:     "0",                            // TODO: This should be configurable. Read from config.Debug()
	}
}

func (c RunnerContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithEnvVariable("RUNNER_NAME", c.Name).
			WithEnvVariable("RUNNER_TEMP", c.Temp).
			WithEnvVariable("RUNNER_OS", c.OS).
			WithEnvVariable("RUNNER_ARCH", c.Arch).
			WithEnvVariable("RUNNER_TOOL_CACHE", c.ToolCache).
			WithMountedCache(c.ToolCache, config.Client().CacheVolume("RUNNER_TOOL_CACHE")).
			WithEnvVariable("RUNNER_DEBUG", c.Debug)
	}
}

var _ helpers.WithContainerFuncHook = new(GithubRepositoryContext)

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

func (c GithubRepositoryContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
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
}

var _ helpers.WithContainerFuncHook = new(GithubSecretsContext)

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

func (c GithubSecretsContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.WithSecretVariable("GITHUB_TOKEN", config.Client().SetSecret("GITHUB_TOKEN", c.Token))
	}
}

var _ helpers.WithContainerFuncHook = new(GithubURLContext)

// GithubURLContext is a context that contains URLs for the Github server and API.
type GithubURLContext struct {
	//nolint:revive,stylecheck // ApiURL is more readable than APIURL
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

func (c GithubURLContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithEnvVariable("GITHUB_API_URL", c.ApiURL).
			WithEnvVariable("GITHUB_GRAPHQL_URL", c.GraphqlURL).
			WithEnvVariable("GITHUB_SERVER_URL", c.ServerURL)
	}
}

// GithubFilesContext is a context that contains paths for files and directories useful to Github Actions.
//
// This context changes per step. It'll be used on ghx level to create a temporary file that sets environment variables
// from workflow commands. It won't be used or applied to the container.
type GithubFilesContext struct {
	Env  string `json:"env"`  // Env is the path to a temporary file that sets environment variables from workflow commands.
	Path string `json:"path"` // Path is the path to a temporary file that sets the system PATH variable from workflow commands.
}

var _ helpers.WithContainerFuncHook = new(GithubWorkflowContext)

// GithubWorkflowContext is a context that contains information about the workflow.
type GithubWorkflowContext struct {
	Workflow      string `json:"workflow"`       // Workflow is the name of the workflow. If the workflow file doesn't specify a name, the value of this property is the full path of the workflow file in the repository.
	WorkflowRef   string `json:"workflow_ref"`   // WorkflowRef is the ref path to the workflow. For example, octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch.
	WorkflowSHA   string `json:"workflow_sha"`   // WorkflowSHA is the commit SHA for the workflow file.
	RunID         string `json:"run_id"`         // RunID is a unique number for each workflow run within a repository. This number does not change if you re-run the workflow run.
	RunNumber     string `json:"run_number"`     // RunNumber is a unique number for each run of a particular workflow in a repository. This number begins at 1 for the workflow's first run, and increments with each new run. This number does not change if you re-run the workflow run.
	RunAttempt    string `json:"run_attempt"`    // RunAttempt is a unique number for each attempt of a particular workflow run in a repository. This number begins at 1 for the workflow run's first attempt, and increments with each re-run.
	RetentionDays string `json:"retention_days"` // RetentionDays is the number of days that workflow run logs and artifacts are kept.
}

// NewGithubWorkflowContext creates a new GithubWorkflowContext from the given workflow.
func NewGithubWorkflowContext(repo *Repository, workflow *Workflow, runID string) GithubWorkflowContext {
	return GithubWorkflowContext{
		Workflow:      workflow.Name,
		WorkflowRef:   fmt.Sprintf("%s/%s@%s", repo.NameWithOwner, workflow.Path, repo.CurrentRef),
		WorkflowSHA:   workflow.SHA,
		RunID:         runID,
		RunNumber:     "1", // TODO: fill this value
		RunAttempt:    "1", // TODO: fill this value
		RetentionDays: "0", // TODO: fill this value
	}
}

func (c GithubWorkflowContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithEnvVariable("GITHUB_WORKFLOW", c.Workflow).
			WithEnvVariable("GITHUB_WORKFLOW_REF", c.WorkflowRef).
			WithEnvVariable("GITHUB_WORKFLOW_SHA", c.WorkflowSHA).
			WithEnvVariable("GITHUB_RUN_ID", c.RunID).
			WithEnvVariable("GITHUB_RUN_NUMBER", c.RunNumber).
			WithEnvVariable("GITHUB_RUN_ATTEMPT", c.RunAttempt).
			WithEnvVariable("GITHUB_RETENTION_DAYS", c.RetentionDays)
	}
}

var _ helpers.WithContainerFuncHook = new(GithubJobInfoContext)

// GithubJobInfoContext is a context that contains information about the job.
type GithubJobInfoContext struct {
	Job string `json:"job"` // Job is the job_id of the current job. Note: This context property is set by the Actions runner, and is only available within the execution steps of a job. Otherwise, the value of this property will be null.

	// TODO: enable these fields when reusable workflows are supported.
	// JobWorkflowSHA string // JobWorkflowSHA is for jobs using a reusable workflow, the commit SHA for the reusable workflow file.
}

func NewGithubJobInfoContext(jobID string) GithubJobInfoContext {
	return GithubJobInfoContext{Job: jobID}
}

func (c GithubJobInfoContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.WithEnvVariable("GITHUB_JOB", c.Job)
	}
}

// JobContext contains information about the currently running job.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
type JobContext struct {
	Status Conclusion `json:"status"` // Status is the current status of the job. Possible values are success, failure, or cancelled.

	// TODO: add other fields when needed.
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

var _ helpers.WithContainerFuncHook = new(SecretsContext)

// SecretsContext is a context that contains secrets.
type SecretsContext map[string]string

// NewSecretsContext creates a new SecretsContext from the given secrets.
func NewSecretsContext(token string, secrets map[string]string) SecretsContext {
	if secrets == nil {
		secrets = make(map[string]string)
	}

	secrets["GITHUB_TOKEN"] = token // GITHUB_TOKEN is a special secret that is always available to the workflow.

	return secrets
}

func (c SecretsContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		data, err := json.Marshal(c)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		secret := config.Client().SetSecret("secrets-context", string(data))

		return container.WithMountedSecret(filepath.Join(config.GhxHome(), "secrets", "secrets.json"), secret)
	}
}
