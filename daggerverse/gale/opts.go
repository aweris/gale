package main

// FIXME: currently Object types can't be used as a type parameter. So, we can't use WorkflowOpts as a type parameter
//  refactor this code when dagger supports Opt struct types as type parameters.
//  issue: https://github.com/dagger/dagger/issues/6162

type WorkflowListOpts struct {
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	Source *Directory

	// The name of the repository. Format: owner/name.
	Repo string

	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Tag string

	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Branch string

	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	WorkflowsDir string
}

// toWorkflowListOpts converts the given options to WorkflowListOpts.
func toWorkflowListOpts(source Optional[*Directory], repo, tag, branch, workflowsDir Optional[string]) *WorkflowListOpts {
	return &WorkflowListOpts{
		Source:       withEmptyValue(source),
		Repo:         withEmptyValue(repo),
		Tag:          withEmptyValue(tag),
		Branch:       withEmptyValue(branch),
		WorkflowsDir: workflowsDir.GetOr(".github/workflows"),
	}
}

type WorkflowRunOpts struct {
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	Source *Directory

	// The name of the repository. Format: owner/name.
	Repo string

	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Tag string

	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	Branch string

	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	WorkflowsDir string

	// External workflow file to run.
	WorkflowFile *File

	// Name of the workflow to run.
	Workflow string

	// Name of the job to run. If empty, all jobs will be run.
	Job string

	// Container to run the workflow in.
	Container *Container

	// Name of the event that triggered the workflow. e.g. push
	Event string

	// File with the complete webhook event payload.
	EventFile *File

	// Enables debug mode.
	RunnerDebug bool

	// GitHub token to use for authentication.
	Token *Secret
}

// toWorkflowRunOpts converts the given options to WorkflowRunOpts.
func toWorkflowRunOpts(
	source Optional[*Directory],
	repo Optional[string],
	tag Optional[string],
	branch Optional[string],
	workflowsDir Optional[string],
	workflowFile Optional[*File],
	workflow Optional[string],
	job Optional[string],
	container Optional[*Container],
	event Optional[string],
	eventFile Optional[*File],
	runnerDebug Optional[bool],
	token Optional[*Secret],
) *WorkflowRunOpts {
	return &WorkflowRunOpts{
		Source:       withEmptyValue(source),
		Repo:         withEmptyValue(repo),
		Tag:          withEmptyValue(tag),
		Branch:       withEmptyValue(branch),
		WorkflowsDir: workflowsDir.GetOr(".github/workflows"),
		WorkflowFile: withEmptyValue(workflowFile),
		Workflow:     withEmptyValue(workflow),
		Job:          withEmptyValue(job),
		Container:    withEmptyValue(container),
		Event:        event.GetOr("push"),
		EventFile:    eventFile.GetOr(dag.Directory().WithNewFile("event.json", "{}").File("event.json")),
		RunnerDebug:  withEmptyValue(runnerDebug),
		Token:        withEmptyValue(token),
	}
}
