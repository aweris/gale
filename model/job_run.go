package model

// JobRunEnv contains the environment variables for a job run.
type JobRunEnv struct {
	WorkflowEnv map[string]string `json:"workflow"` // Workflow is the environment variables for the workflow.
	JobEnv      map[string]string `json:"job"`      // Job is the environment variables for the job.
}

// JobRunContext contains the context for a job run.
type JobRunContext struct {
	Github *GithubContext `json:"github"` // Github is the github context used in job run
	Runner *RunnerContext `json:"runner"` // Runner is the runner context used in job run
}
