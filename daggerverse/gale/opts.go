package main

// RepoOpts represents the options for getting repository information.
//
// This is copy of RepoOpts from daggerverse/gale/repo.go to be able to expose options with gale module and pass them to
// the repo module just type casting.
type RepoOpts struct {
	Source *Directory `doc:"Directory containing the repository source. Has precedence over repo."`
	Repo   string     `doc:"Name of the repository. Format: owner/name."`
	Branch string     `doc:"Git branch to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence."`
	Tag    string     `doc:"Git tag to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence."`
}

// WorkflowsDirOpts represents the options for getting workflow information.
type WorkflowsDirOpts struct {
	WorkflowsDir string `doc:"Path to the workflow directory." default:".github/workflows"`
}

// WorkflowsRunOpts represents the options for running a workflow.
type WorkflowsRunOpts struct {
	WorkflowFile *File   `doc:"External workflow file to run."`
	Workflow     string  `doc:"Name of the workflow to run."`
	Job          string  `doc:"Name of the job to run. If empty, all jobs will be run."`
	Event        string  `doc:"Name of the event that triggered the workflow." default:"push"`
	EventFile    *File   `doc:"The file with the complete webhook json event payload."`
	RunnerImage  string  `doc:"Docker image to use for the runner." default:"ghcr.io/catthehacker/ubuntu:act-latest"`
	RunnerDebug  bool    `doc:"Enables debug mode." default:"false"`
	Token        *Secret `doc:"GitHub token to use for authentication."`
}

// WorkflowRunDirectoryOpts represents the options for exporting a workflow run.
type WorkflowRunDirectoryOpts struct {
	IncludeRepo      bool `doc:"Adds the repository source to the exported directory." default:"false"`
	IncludeSecrets   bool `doc:"Adds the mounted secrets to the exported directory." default:"false"`
	IncludeEvent     bool `doc:"Adds the event file to the exported directory." default:"false"`
	IncludeArtifacts bool `doc:"Adds the uploaded artifacts to the exported directory." default:"false"`
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

	// WorkflowFile is external workflow file to run.
	WorkflowFile *File

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
