package core

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"dagger.io/dagger"

	"github.com/google/uuid"

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
func NewRunnerContext(debug bool) RunnerContext {
	// adjust debug value from bool to string as it's expected to be string in the container.
	debugVal := "0"
	if debug {
		debugVal = "1"
	}
	return RunnerContext{
		Name:      "Gale Agent",
		OS:        "linux",
		Arch:      "x64",
		Temp:      "/home/runner/_temp",
		ToolCache: "/home/runner/hostedtoolcache", // /opt/hostedtoolcache is used by our base runner image and if we mount it we'll lose the tools installed by the base image.
		Debug:     debugVal,
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
		Dir:               repo.GitRef.Dir,
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
func NewGithubWorkflowContext(repo *Repository, workflow *Workflow, runID, workflowSHA string) GithubWorkflowContext {
	return GithubWorkflowContext{
		Workflow:      workflow.Name,
		WorkflowRef:   fmt.Sprintf("%s/%s@%s", repo.NameWithOwner, workflow.Path, repo.GitRef.Ref),
		WorkflowSHA:   workflowSHA,
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

// GithubRefContext is a context that contains information about the Git ref that triggered the workflow.
type GithubRefContext struct {
	Ref          string `json:"ref"`           // Ref is the branch or tag ref that triggered the workflow. If neither a branch or tag is available for the event type, the variable will not exist.
	RefName      string `json:"ref_name"`      // RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow. If neither a branch or tag is available for the event type, the variable will not exist.
	RefType      string `json:"ref_type"`      // RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither a branch or tag is available for the event type.
	RefProtected bool   `json:"ref_protected"` // RefProtected is true if branch protections are enabled and the base ref for the pull request matches the branch protection rule.
	BaseRef      string `json:"base_ref"`      // BaseRef is the branch of the base repository. This property is only available when the event that triggered the workflow is a pull request. Otherwise the property will not exist.
	HeadRef      string `json:"head_ref"`      // HeadRef is the branch of the head repository. This property is only available when the event that triggered the workflow is a pull request. Otherwise the property will not exist.
	SHA          string `json:"sha"`           // SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that triggered the workflow.
}

// NewGithubRefContext creates a new GithubRefContext from the given repository and ref.
func NewGithubRefContext(ref *RepositoryGitRef) GithubRefContext {
	return GithubRefContext{
		Ref:          ref.Ref,
		RefName:      ref.RefName,
		RefType:      string(ref.RefType),
		RefProtected: false, // TODO: fill this value when needed, not supported yet.
		BaseRef:      "",    // TODO: fill this value when needed, not supported yet.
		HeadRef:      "",    // TODO: fill this value when needed, not supported yet.
		SHA:          ref.SHA,
	}
}

func (c GithubRefContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.
			WithEnvVariable("GITHUB_REF", c.Ref).
			WithEnvVariable("GITHUB_REF_NAME", c.RefName).
			WithEnvVariable("GITHUB_REF_TYPE", c.RefType).
			WithEnvVariable("GITHUB_REF_PROTECTED", strconv.FormatBool(c.RefProtected)).
			WithEnvVariable("GITHUB_BASE_REF", c.BaseRef).
			WithEnvVariable("GITHUB_HEAD_REF", c.HeadRef).
			WithEnvVariable("GITHUB_SHA", c.SHA)
	}
}

// GithubEventContext is a context that contains information about the event that triggered the workflow.
type GithubEventContext struct {
	EventName string                 `json:"event_name"` // EventName is the name of the event that triggered the workflow. e.g. push
	EventPath string                 `json:"event_path"` // EventPath is the path of the file with the complete webhook event payload. e.g. /github/workflow/event.json
	Event     map[string]interface{} `json:"event"`      // Event is the full event webhook payload.
}

// NewGithubEventContext creates a new GithubEventContext from the given event.
func NewGithubEventContext() GithubEventContext {
	return GithubEventContext{
		EventName: "push", // TODO: this is only supported event type for now. Make it configurable when we support events properly.
		EventPath: filepath.Join("/home/runner/_temp", uuid.NewString(), "event.json"),
		Event:     make(map[string]interface{}),
	}
}

func (c GithubEventContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		data, err := json.Marshal(c.Event)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		return container.
			WithEnvVariable("GITHUB_EVENT_NAME", c.EventName).
			WithEnvVariable("GITHUB_EVENT_PATH", c.EventPath).
			WithNewFile(c.EventPath, dagger.ContainerWithNewFileOpts{Contents: string(data)})
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
