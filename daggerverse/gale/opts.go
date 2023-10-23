package main

// RepoOpts represents the options for getting repository information.
//
// This is copy of RepoOpts from daggerverse/gale/repo.go to be able to expose options with gale module and pass them to
// the repo module just type casting.
type RepoOpts struct {
	Source *Directory `doc:"The directory containing the repository source. If source is provided, rest of the options are ignored."`
	Repo   string     `doc:"The name of the repository. Format: owner/name."`
	Branch string     `doc:"Branch name to checkout. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
	Tag    string     `doc:"Tag name to checkout. Only one of branch or tag can be used. Precedence is as follows: tag, branch."`
}

// WorkflowsDirOpts represents the options for getting workflow information.
type WorkflowsDirOpts struct {
	WorkflowsDir string `doc:"The relative path to the workflow directory." default:".github/workflows"`
}

// WorkflowsRunOpts represents the options for running a workflow.
type WorkflowsRunOpts struct {
	Workflow    string  `doc:"The workflow to run." required:"true"`
	Job         string  `doc:"The job name to run. If empty, all jobs will be run."`
	Event       string  `doc:"Name of the event that triggered the workflow. e.g. push" default:"push"`
	EventFile   *File   `doc:"The file with the complete webhook event payload."`
	RunnerImage string  `doc:"The image to use for the runner." default:"ghcr.io/catthehacker/ubuntu:act-latest"`
	RunnerDebug bool    `doc:"Enable debug mode." default:"false"`
	Token       *Secret `doc:"The GitHub token to use for authentication."`
}

// WorkflowRunDirectoryOpts represents the options for exporting a workflow run.
type WorkflowRunDirectoryOpts struct {
	IncludeRepo      bool `doc:"Include the repository source in the exported directory." default:"false"`
	IncludeSecrets   bool `doc:"Include the secrets in the exported directory." default:"false"`
	IncludeEvent     bool `doc:"Include the event file in the exported directory." default:"false"`
	IncludeArtifacts bool `doc:"Include the artifacts in the exported directory." default:"false"`
}

// FIXME: add jobs to WorkflowRunReport when dagger supports map type

// WorkflowRunReport represents the result of a workflow run.
type WorkflowRunReport struct {
	Ran           bool   `json:"ran"`            // Ran indicates if the execution ran
	Duration      string `json:"duration"`       // Duration of the execution
	Name          string `json:"name"`           // Name is the name of the workflow
	Path          string `json:"path"`           // Path is the path of the workflow
	RunID         string `json:"run_id"`         // RunID is the ID of the run
	RunNumber     string `json:"run_number"`     // RunNumber is the number of the run
	RunAttempt    string `json:"run_attempt"`    // RunAttempt is the attempt number of the run
	RetentionDays string `json:"retention_days"` // RetentionDays is the number of days to keep the run logs
	Conclusion    string `json:"conclusion"`     // Conclusion is the result of a completed workflow run after continue-on-error is applied
}

// Configuration objects

// When dagger converts objects to commands. All public methods and fields converted to sub commands of that object.
// To keep command DX simple when we need to pass options as a part of the state we're using config objects to hold the
// options. This way we can keep the command DX simple and only one extra command is added to the command tree.

// These config objects basically holds the options for the commands but to keep code code gen friendly we're duplicating
// the options as fields of the config objects.

// WorkflowRunConfig holds the configuration of a workflow run.
type WorkflowRunConfig struct {
	// Directory containing the repository source.
	Source *Directory

	// Name of the repository. Format: owner/name.
	Repo string

	// Branch name to check out. Only one of branch or tag can be used. Precedence: tag, branch.
	Branch string

	// Tag name to check out. Only one of branch or tag can be used. Precedence: tag, branch.
	Tag string

	// Path to the workflow directory.
	WorkflowsDir string

	// Workflow to run.
	Workflow string

	// Job name to run. If empty, all jobs will be run.
	Job string

	// Name of the event that triggered the workflow. e.g. push
	Event string

	// File with the complete webhook event payload.
	EventFile *File

	// Image to use for the runner.
	RunnerImage string

	// Enables debug mode.
	RunnerDebug bool

	// GitHub token to use for authentication.
	Token *Secret
}
