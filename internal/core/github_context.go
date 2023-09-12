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

var _ helpers.WithContainerFuncHook = new(GithubContext)

type GithubContext struct {
	Repository        string                 `json:"repository"`          // Repository is the combination of owner and name of the repository. e.g. octocat/hello-world
	RepositoryID      string                 `json:"repository_id"`       // RepositoryID is the id of the repository. e.g. 1296269. Note that this is different from the repository name.
	RepositoryOwner   string                 `json:"repository_owner"`    // RepositoryOwner is the owner of the repository. e.g. octocat
	RepositoryOwnerID string                 `json:"repository_owner_id"` // RepositoryOwnerID is the id of the repository owner. e.g. 1234567. Note that this is different from the repository owner name.
	RepositoryURL     string                 `json:"repository_url"`      // RepositoryURL is the git url of the repository. e.g. git://github.com/octocat/hello-world.git.
	Workspace         string                 `json:"workspace"`           // Workspace is the path of a directory that contains a checkout of the repository.
	APIURL            string                 `json:"api_url"`             // ApiURL is the URL of the Github API. e.g. https://api.github.com
	GraphqlURL        string                 `json:"graphql_url"`         // GraphqlURL is the URL of the Github GraphQL API. e.g. https://api.github.com/graphql
	ServerURL         string                 `json:"server_url"`          // ServerURL is the URL of the Github server. e.g. https://github.com
	Env               string                 `json:"env"`                 // Env is the path to a temporary file that sets environment variables from workflow commands.
	Path              string                 `json:"path"`                // Path is the path to a temporary file that sets the system PATH variable from workflow commands.
	Workflow          string                 `json:"workflow"`            // Workflow is the name of the workflow. If the workflow file doesn't specify a name, the value of this property is the full path of the workflow file in the repository.
	WorkflowRef       string                 `json:"workflow_ref"`        // WorkflowRef is the ref path to the workflow. For example, octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch.
	WorkflowSHA       string                 `json:"workflow_sha"`        // WorkflowSHA is the commit SHA for the workflow file.
	Job               string                 `json:"job"`                 // Job is the job_id of the current job. Note: This context property is set by the Actions runner, and is only available within the execution steps of a job. Otherwise, the value of this property will be null.
	JobWorkflowSHA    string                 `json:"job_workflow_sha"`    // JobWorkflowSHA for jobs using a reusable workflow, the commit SHA for the reusable workflow file.
	RunID             string                 `json:"run_id"`              // RunID is a unique number for each workflow run within a repository. This number does not change if you re-run the workflow run.
	RunNumber         string                 `json:"run_number"`          // RunNumber is a unique number for each run of a particular workflow in a repository. This number begins at 1 for the workflow's first run, and increments with each new run. This number does not change if you re-run the workflow run.
	RunAttempt        string                 `json:"run_attempt"`         // RunAttempt is a unique number for each attempt of a particular workflow run in a repository. This number begins at 1 for the workflow run's first attempt, and increments with each re-run.
	RetentionDays     string                 `json:"retention_days"`      // RetentionDays is the number of days that workflow run logs and artifacts are kept.
	Ref               string                 `json:"ref"`                 // Ref is the branch or tag ref that triggered the workflow. If neither a branch or tag is available for the event type, the variable will not exist.
	RefName           string                 `json:"ref_name"`            // RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow. If neither a branch or tag is available for the event type, the variable will not exist.
	RefType           string                 `json:"ref_type"`            // RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither a branch or tag is available for the event type.
	RefProtected      bool                   `json:"ref_protected"`       // RefProtected is true if branch protections are enabled and the base ref for the pull request matches the branch protection rule.
	BaseRef           string                 `json:"base_ref"`            // BaseRef is the branch of the base repository. This property is only available when the event that triggered the workflow is a pull request. Otherwise the property will not exist.
	HeadRef           string                 `json:"head_ref"`            // HeadRef is the branch of the head repository. This property is only available when the event that triggered the workflow is a pull request. Otherwise the property will not exist.
	SHA               string                 `json:"sha"`                 // SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that triggered the workflow.
	EventName         string                 `json:"event_name"`          // EventName is the name of the event that triggered the workflow. e.g. push
	EventPath         string                 `json:"event_path"`          // EventPath is the path of the file with the complete webhook event payload. e.g. /github/workflow/event.json
	Event             map[string]interface{} `json:"event"`               // Event is the full event webhook payload.
	Token             string                 `json:"token"`               // Token is the GitHub token to use for authentication.
}

func NewGithubContext(repo *Repository, token string) *GithubContext {
	gc := &GithubContext{}

	gc.SetDefaults()
	gc.SetRepo(repo)
	gc.SetToken(token)

	return gc
}

// SetDefaults sets the default values for the context.
//
// TODO: replace these properties with actual values. These are just placeholders to have some values for now.
func (c *GithubContext) SetDefaults() *GithubContext {
	// Github API
	c.APIURL = "https://api.github.com"
	c.GraphqlURL = "https://api.github.com/graphql"
	c.ServerURL = "https://github.com"

	// Repository
	c.RefProtected = false
	c.BaseRef = ""
	c.HeadRef = ""

	// Event
	c.EventName = "push"
	c.EventPath = filepath.Join("/home/runner/_temp", uuid.NewString(), "event.json")
	c.Event = make(map[string]interface{})

	return c
}

// SetRepo sets the repository information in the context.
func (c *GithubContext) SetRepo(repo *Repository) *GithubContext {
	c.Repository = repo.NameWithOwner
	c.RepositoryID = repo.ID
	c.RepositoryOwner = repo.Owner.Login
	c.RepositoryOwnerID = repo.Owner.ID
	c.RepositoryURL = repo.URL
	c.Workspace = fmt.Sprintf("/home/runner/work/%s/%s", repo.Name, repo.Name)

	ref := repo.GitRef

	c.Ref = ref.Ref
	c.RefName = ref.RefName
	c.RefType = string(ref.RefType)
	c.SHA = ref.SHA

	return c
}

// SetJob sets the job information in the context.
func (c *GithubContext) SetJob(job *Job) *GithubContext {
	c.Job = job.ID

	return c
}

func (c *GithubContext) SetWorkflowRun(run *WorkflowRun) *GithubContext {
	c.RunID = run.RunID
	c.RunNumber = run.RunNumber
	c.RunAttempt = run.RunAttempt
	c.RetentionDays = run.RetentionDays
	c.Workflow = run.Workflow.Name
	c.WorkflowRef = fmt.Sprintf("%s/%s@%s", c.Repository, run.Workflow.Path, c.Ref)
	c.WorkflowSHA = c.Ref // TODO: double check this. It should be ok for know since we're loading the workflow from the repository only.

	return c
}

// SetToken sets the token in the context.
func (c *GithubContext) SetToken(token string) *GithubContext {
	c.Token = token

	return c
}

// WithContainerFunc returns a WithContainerFunc that sets the context in the container.
func (c *GithubContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		event, err := json.Marshal(c.Event)
		if err != nil {
			helpers.FailPipeline(container, err)
		}

		return container.
			WithEnvVariable("GITHUB_REPOSITORY", c.Repository).
			WithEnvVariable("GITHUB_REPOSITORY_ID", c.RepositoryID).
			WithEnvVariable("GITHUB_REPOSITORY_OWNER", c.RepositoryOwner).
			WithEnvVariable("GITHUB_REPOSITORY_OWNER_ID", c.RepositoryOwnerID).
			WithEnvVariable("GITHUB_REPOSITORY_URL", c.RepositoryURL).
			WithEnvVariable("GITHUB_WORKSPACE", c.Workspace).
			WithEnvVariable("GITHUB_API_URL", c.APIURL).
			WithEnvVariable("GITHUB_GRAPHQL_URL", c.GraphqlURL).
			WithEnvVariable("GITHUB_SERVER_URL", c.ServerURL).
			WithEnvVariable("GITHUB_REF", c.Ref).
			WithEnvVariable("GITHUB_REF_NAME", c.RefName).
			WithEnvVariable("GITHUB_REF_TYPE", c.RefType).
			WithEnvVariable("GITHUB_REF_PROTECTED", strconv.FormatBool(c.RefProtected)).
			WithEnvVariable("GITHUB_BASE_REF", c.BaseRef).
			WithEnvVariable("GITHUB_HEAD_REF", c.HeadRef).
			WithEnvVariable("GITHUB_SHA", c.SHA).
			WithEnvVariable("GITHUB_EVENT_NAME", c.EventName).
			WithEnvVariable("GITHUB_EVENT_PATH", c.EventPath).
			WithNewFile(c.EventPath, dagger.ContainerWithNewFileOpts{Contents: string(event)}).
			WithSecretVariable("GITHUB_TOKEN", config.Client().SetSecret("GITHUB_TOKEN", c.Token))
	}
}
