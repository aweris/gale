package context

import (
	"dagger.io/dagger"

	"github.com/aweris/gale/common/model"
)

type GhxConfig struct {
	// Workflow name to run.
	Workflow string `env:"GHX_WORKFLOW"`

	// Job name to run. If not specified, the all jobs will be run.
	Job string `env:"GHX_JOB"`

	// Home directory for the ghx to use for storing execution related files.
	HomeDir string `env:"GHX_HOME" envDefault:"/home/runner/_temp/ghx"`

	// ActionsDir is the directory to look for actions.
	ActionsDir string `env:"GHX_ACTIONS_DIR" envDefault:"/home/runner/_temp/gale/actions"`

	// MetadataDir is the directory to look for metadata.
	MetadataDir string `env:"GHX_METADATA_DIR" envDefault:"/home/runner/_temp/gale/metadata"`
}

// DaggerContext is the context holding the dagger client.
type DaggerContext struct {
	// Client is the dagger client to be used in the workflow.
	Client *dagger.Client
}

type ExecutionContext struct {
	// Workflow is the current workflow that is being executed.
	WorkflowRun *model.WorkflowRun

	// Job is the current job that is being executed.
	JobRun *model.JobRun

	// Step is the current step that is being executed.
	StepRun *model.StepRun

	// CurrentAction is the current action that is being executed. This is only available on step level if the step is uses a custom action.
	CurrentAction *model.CustomAction
}

// ActionsContext is the context for the internal services configuration for used by GitHub Actions.
type ActionsContext struct {
	// RuntimeURL is the URL for the actions runtime. In scope of gale, this is the URL of the artifact service.
	RuntimeURL string `env:"ACTIONS_RUNTIME_URL"`

	// CacheURL is the URL for the actions cache service.
	CacheURL string `env:"ACTIONS_CACHE_URL"`

	// Token is the token for the actions runtime. In scope of gale, this is a dummy token.
	Token string `env:"ACTIONS_RUNTIME_TOKEN" envDefault:"dummy-token"`
}

// GithubContext contains information about the workflow run and the event that triggered the run.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type GithubContext struct {
	// Repository is the combination of owner and name of the repository. e.g. octocat/hello-world
	Repository string `json:"repository" env:"GITHUB_REPOSITORY"`

	// RepositoryID is the id of the repository. e.g. 1296269. Note that this is different from the repository name.
	RepositoryID string `json:"repository_id" env:"GITHUB_REPOSITORY_ID"`

	// RepositoryOwner is the owner of the repository. e.g. octocat
	RepositoryOwner string `json:"repository_owner" env:"GITHUB_REPOSITORY_OWNER"`

	// RepositoryOwnerID is the id of the repository owner. e.g. 1234567. Note that this is different from
	// the repository owner name.
	RepositoryOwnerID string `json:"repository_owner_id" env:"GITHUB_REPOSITORY_OWNER_ID"`

	// RepositoryURL is the git url of the repository. e.g. git://github.com/octocat/hello-world.git.
	RepositoryURL string `json:"repository_url" env:"GITHUB_REPOSITORY_URL"`

	// Workspace is the path of a directory that contains a checkout of the repository.
	Workspace string `json:"workspace" env:"GITHUB_WORKSPACE"`

	// ApiURL is the CloneURL of the Github API. e.g. https://api.github.com
	APIURL string `json:"api_url" env:"GITHUB_API_URL" envDefault:"https://api.github.com"`

	// GraphqlURL is the CloneURL of the Github GraphQL API. e.g. https://api.github.com/graphql
	GraphqlURL string `json:"graphql_url" env:"GITHUB_GRAPHQL_URL" envDefault:"https://api.github.com/graphql"`

	// ServerURL is the CloneURL of the Github server. e.g. https://github.com
	ServerURL string `json:"server_url" env:"GITHUB_SERVER_URL" envDefault:"https://github.com"`

	// Env is the path to a temporary file that sets environment variables from workflow commands.
	Env string `json:"env" env:"GITHUB_ENV"`

	// Path is the path to a temporary file that sets the system PATH variable from workflow commands.
	Path string `json:"path" env:"GITHUB_PATH"`

	// Workflow is the name of the workflow. If the workflow file doesn't specify a name, the value of this property
	// is the full path of the workflow file in the repository.
	Workflow string `json:"workflow" env:"GITHUB_WORKFLOW"`

	// WorkflowRef is the ref path to the workflow. For example, octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch.
	WorkflowRef string `json:"workflow_ref" env:"GITHUB_WORKFLOW_REF"`

	// WorkflowSHA is the commit SHA for the workflow file.
	WorkflowSHA string `json:"workflow_sha" env:"GITHUB_WORKFLOW_SHA"`

	// Job is the job_id of the current job. Note: This context property is set by the Actions runner, and is
	// only available within the execution steps of a job. Otherwise, the value of this property will be null.
	Job string `json:"job" env:"GITHUB_JOB"`

	// JobWorkflowSHA for jobs using a reusable workflow, the commit SHA for the reusable workflow file.
	JobWorkflowSHA string `json:"job_workflow_sha" env:"GITHUB_JOB_WORKFLOW_SHA"`

	// RunID is a unique number for each workflow run within a repository. This number does not change if you re-run
	// the workflow run.
	RunID string `json:"run_id" env:"GITHUB_RUN_ID"`

	// RunNumber is a unique number for each run of a particular workflow in a repository. This number begins at 1
	// for the workflow's first run, and increments with each new run. This number does not change if you re-run the workflow run.
	RunNumber string `json:"run_number" env:"GITHUB_RUN_NUMBER"`

	// RunAttempt is a unique number for each attempt of a particular workflow run in a repository. This number
	// begins at 1 for the workflow run's first attempt, and increments with each re-run.
	RunAttempt string `json:"run_attempt" env:"GITHUB_RUN_ATTEMPT"`

	// RetentionDays is the number of days that workflow run logs and artifacts are kept.
	RetentionDays string `json:"retention_days" env:"GITHUB_RETENTION_DAYS"`

	// Ref is the branch or tag ref that triggered the workflow. If neither a branch or tag is available for
	// the event type, the variable will not exist.
	Ref string `json:"ref" env:"GITHUB_REF"  `

	// RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow.
	// If neither a branch or tag is available for the event type, the variable will not exist.
	RefName string `json:"ref_name" env:"GITHUB_REF_NAME"`

	// RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither
	// a branch nor tag is available for the event type.
	RefType string `json:"ref_type" env:"GITHUB_REF_TYPE"`

	// RefProtected is true if branch protections are enabled and the base ref for the pull request matches the branch
	// protection rule.
	RefProtected bool `json:"ref_protected" env:"GITHUB_REF_PROTECTED"`

	// BaseRef is the branch of the base repository. This property is only available when the event that triggered
	// the workflow is a pull request. Otherwise the property will not exist.
	BaseRef string `json:"base_ref" env:"GITHUB_BASE_REF"`

	// HeadRef is the branch of the head repository. This property is only available when the event that triggered
	// the workflow is a pull request. Otherwise the property will not exist.
	HeadRef string `json:"head_ref" env:"GITHUB_HEAD_REF"`

	// SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	// triggered the workflow.
	SHA string `json:"sha" env:"GITHUB_SHA"`

	// EventName is the name of the event that triggered the workflow. e.g. push
	EventName string `json:"event_name" env:"GITHUB_EVENT_NAME" envDefault:"push"`

	// EventPath is the path of the file with the complete webhook event payload. e.g. /github/workflow/event.json
	EventPath string `json:"event_path" env:"GITHUB_EVENT_PATH"`

	// Event is the full event webhook payload.
	Event map[string]interface{} `json:"event"`

	// Token is the GitHub token to use for authentication.
	Token string `json:"token" env:"GITHUB_TOKEN"`
}

// InputsContext contains input properties passed to an action, to a reusable workflow, or to a manually triggered
// workflow.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#inputs-context
type InputsContext map[string]string

// JobContext contains information about the currently running job.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#job-context
type JobContext struct {
	Status model.Conclusion `json:"status"` // Status is the current status of the job. Possible values are success, failure, or cancelled.

	// TODO: add other fields when needed.
}

// NeedsContext is a context that contains information about dependent job.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#needs-context
type NeedsContext map[string]NeedContext

// NeedContext is a context that contains information about dependent job.
type NeedContext struct {
	Result  model.Conclusion  `json:"result"`  // Result is conclusion of the job. It can be success, failure, skipped or cancelled.
	Outputs map[string]string `json:"outputs"` // Outputs of the job
}

// RunnerContext contains information about the runner environment.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#runner-context
type RunnerContext struct {
	// Name is the name of the runner.
	Name string `json:"name" env:"RUNNER_NAME" envDefault:"Gale Agent"`

	// OS is the operating system of the runner.
	OS string `json:"os" env:"RUNNER_OS" envDefault:"linux"`

	// Arch is the architecture of the runner.
	Arch string `json:"arch" env:"RUNNER_ARCH" envDefault:"x64"`

	// Temp is the path to the directory containing temporary files created by the runner during the job.
	Temp string `json:"temp" env:"RUNNER_TEMP" envDefault:"/home/runner/_temp"`

	// ToolCache is the path to the directory containing installed tools.
	ToolCache string `json:"tool_cache" env:"RUNNER_TOOL_CACHE" envDefault:"/home/runner/hostedtoolcache"`

	// Debug is a boolean value that indicates whether to run the runner in debug mode.
	Debug string `json:"debug" env:"RUNNER_DEBUG" envDefault:"0"`
}

// SecretsContext is a context that contains secrets.
type SecretsContext struct {
	// MountPath is the path where the secrets are mounted.
	MountPath string

	// Data is the secrets data.
	Data map[string]string
}

// StepsContext is a context that contains information about the steps.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#steps-context
type StepsContext map[string]StepContext

// StepContext is a context that contains information about single step.
//
// This context created per step execution.
type StepContext struct {
	// Conclusion is the result of a completed step after continue-on-error is applied
	Conclusion model.Conclusion `json:"conclusion"`

	// Outcome is  the result of a completed step before continue-on-error is applied
	Outcome model.Conclusion `json:"outcome"`

	// Outputs is a map of output name to output value
	Outputs map[string]string `json:"outputs"`

	// State is a map of step state variables. This is not available to expressions so that's why json tag is set to "-" to ignore it.
	State map[string]string `json:"-"`

	// Summary is the summary of the step. This is not available to expressions so that's why json tag is set to "-" to ignore it.
	Summary string `json:"-"`
}

// EnvContext is a context that contains environment variables.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#env-context
type EnvContext map[string]string

// MatrixContext is a context that contains matrix information.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#matrix-context
type MatrixContext model.MatrixCombination
