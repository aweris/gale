package gctx

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/google/uuid"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/fs"
)

// GithubContext contains information about the workflow run and the event that triggered the run.
//
// See: https://docs.github.com/en/actions/learn-github-actions/contexts#github-context
type GithubContext struct {
	// Repository is the combination of owner and name of the repository. e.g. octocat/hello-world
	Repository string `json:"repository" env:"GITHUB_REPOSITORY" container_env:"true"`

	// RepositoryID is the id of the repository. e.g. 1296269. Note that this is different from the repository name.
	RepositoryID string `json:"repository_id" env:"GITHUB_REPOSITORY_ID" container_env:"true"`

	// RepositoryOwner is the owner of the repository. e.g. octocat
	RepositoryOwner string `json:"repository_owner" env:"GITHUB_REPOSITORY_OWNER" container_env:"true"`

	// RepositoryOwnerID is the id of the repository owner. e.g. 1234567. Note that this is different from
	// the repository owner name.
	RepositoryOwnerID string `json:"repository_owner_id" env:"GITHUB_REPOSITORY_OWNER_ID" container_env:"true"`

	// RepositoryURL is the git url of the repository. e.g. git://github.com/octocat/hello-world.git.
	RepositoryURL string `json:"repository_url" env:"GITHUB_REPOSITORY_URL" container_env:"true"`

	// Workspace is the path of a directory that contains a checkout of the repository.
	Workspace string `json:"workspace" env:"GITHUB_WORKSPACE" container_env:"true"`

	// ApiURL is the URL of the Github API. e.g. https://api.github.com
	APIURL string `json:"api_url" env:"GITHUB_API_URL" envDefault:"https://api.github.com" container_env:"true"`

	// GraphqlURL is the URL of the Github GraphQL API. e.g. https://api.github.com/graphql
	GraphqlURL string `json:"graphql_url" env:"GITHUB_GRAPHQL_URL" envDefault:"https://api.github.com/graphql" container_env:"true"`

	// ServerURL is the URL of the Github server. e.g. https://github.com
	ServerURL string `json:"server_url" env:"GITHUB_SERVER_URL" envDefault:"https://github.com" container_env:"true"`

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
	Ref string `json:"ref" env:"GITHUB_REF"  container_env:"true"`

	// RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow.
	// If neither a branch or tag is available for the event type, the variable will not exist.
	RefName string `json:"ref_name" env:"GITHUB_REF_NAME" container_env:"true"`

	// RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither
	// a branch nor tag is available for the event type.
	RefType string `json:"ref_type" env:"GITHUB_REF_TYPE" container_env:"true"`

	// RefProtected is true if branch protections are enabled and the base ref for the pull request matches the branch
	// protection rule.
	RefProtected bool `json:"ref_protected" env:"GITHUB_REF_PROTECTED" container_env:"true"`

	// BaseRef is the branch of the base repository. This property is only available when the event that triggered
	// the workflow is a pull request. Otherwise the property will not exist.
	BaseRef string `json:"base_ref" env:"GITHUB_BASE_REF" container_env:"true"`

	// HeadRef is the branch of the head repository. This property is only available when the event that triggered
	// the workflow is a pull request. Otherwise the property will not exist.
	HeadRef string `json:"head_ref" env:"GITHUB_HEAD_REF" container_env:"true"`

	// SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	// triggered the workflow.
	SHA string `json:"sha" env:"GITHUB_SHA" container_env:"true"`

	// EventName is the name of the event that triggered the workflow. e.g. push
	EventName string `json:"event_name" env:"GITHUB_EVENT_NAME" container_env:"true"`

	// EventPath is the path of the file with the complete webhook event payload. e.g. /github/workflow/event.json
	EventPath string `json:"event_path" env:"GITHUB_EVENT_PATH" container_env:"true"`

	// Event is the full event webhook payload.
	Event map[string]interface{} `json:"event"`

	// Token is the GitHub token to use for authentication.
	Token string `json:"token" env:"GITHUB_TOKEN" container_secret:"true"`
}

func (c *Context) LoadGithubContext() error {
	gc, err := NewContextFromEnv[GithubContext]()
	if err != nil {
		return err
	}

	// TODO: replace these event properties more proper values.

	// read event if event path is set
	if gc.EventPath != "" {
		err := fs.ReadJSONFile(gc.EventPath, &gc.Event)
		if err != nil {
			return err
		}
	}

	// Set defaults if not set

	if gc.EventName == "" {
		gc.EventName = "push"
	}

	if gc.Event == nil {
		gc.Event = make(map[string]interface{})
	}

	// Override event path since we already read the event file. We'll write it again in the container with the
	// new event path.
	gc.EventPath = filepath.Join("/home/runner/_temp", uuid.NewString(), "event.json")

	c.Github = gc

	return nil
}

// SetRepo sets the repository information in the context.
func (c *GithubContext) setRepo(repo core.Repository, ref core.RepositoryGitRef) *GithubContext {
	c.Repository = repo.NameWithOwner
	c.RepositoryID = repo.ID
	c.RepositoryOwner = repo.Owner.Login
	c.RepositoryOwnerID = repo.Owner.ID
	c.RepositoryURL = repo.URL
	c.Workspace = fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name)

	c.Ref = ref.Ref
	c.RefName = ref.RefName
	c.RefType = string(ref.RefType)
	c.SHA = ref.SHA

	return c
}

// setToken sets the token in the context.
func (c *GithubContext) setToken(token string) *GithubContext {
	c.Token = token

	return c
}

// setWorkflow sets the workflow information in the context.
func (c *GithubContext) setWorkflow(wr *core.WorkflowRun) *GithubContext {
	c.RunID = wr.RunID
	c.RunNumber = wr.RunNumber
	c.RunAttempt = wr.RunAttempt
	c.RetentionDays = wr.RetentionDays
	c.Workflow = wr.Workflow.Name
	c.WorkflowRef = fmt.Sprintf("%s/%s@%s", c.Repository, wr.Workflow.Path, c.Ref)
	// TODO: double check this. It should be ok for know since we're loading the workflow from the repository only.
	c.WorkflowSHA = c.SHA

	return c
}

// helpers.WithContainerFuncHook interface to be loaded in the container.

var _ helpers.WithContainerFuncHook = new(GithubContext)

// WithContainerFunc returns a WithContainerFunc that sets the context in the container.
func (c *GithubContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		// Load context in the container as environment variables or secrets.
		container = container.With(WithContainerEnv(config.Client(), c))

		// Apply extra container configuration
		event, err := json.Marshal(c.Event)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		container = container.WithNewFile(c.EventPath, dagger.ContainerWithNewFileOpts{Contents: string(event)})

		return container
	}
}
