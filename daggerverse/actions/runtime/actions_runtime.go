package main

type ActionsRuntime struct{}

// example usage: "dagger call container-echo --string-arg yo"
func (m *ActionsRuntime) Run(
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// The action to run. it should be in the format of <action>@<version>
	uses string,
	// Name of the event that triggered the workflow. e.g. push
	event Optional[string],
	// File with the complete webhook event payload.
	eventFile Optional[*File],
	// Image to use for the runner. If --image and --container provided together, --image takes precedence.
	image Optional[string],
	// Container to use for the runner. If --image and --container provided together, --image takes precedence.
	container Optional[*Container],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token to use for authentication.
	token Optional[*Secret],
) *ActionRun {
	return &ActionRun{
		Config: ActionRunConfig{
			Source:      source.GetOr(nil),
			Repo:        repo.GetOr(""),
			Branch:      branch.GetOr(""),
			Tag:         tag.GetOr(""),
			Uses:        uses,
			Env:         []string{},
			With:        []string{},
			Event:       event.GetOr("push"),
			EventFile:   eventFile.GetOr(nil),
			Image:       image.GetOr(""), // default value handled by gale module. Handling it here would override the container if it is provided.
			Container:   container.GetOr(nil),
			RunnerDebug: runnerDebug.GetOr(false),
			Token:       token.GetOr(nil),
		},
	}
}
