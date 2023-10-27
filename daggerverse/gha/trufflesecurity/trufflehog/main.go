// Code generated by gha2dagger. DO NOT EDIT.

package main

// Trufflehog represents the GitHub action. Scan Github Actions with TruffleHog
type Trufflehog struct{}

func (m Trufflehog) Run(
	// Repository path
	withPath string,
	// Start scanning from here (usually main branch).
	withBase Optional[string],
	// Scan commits until here (usually dev branch).
	withHead Optional[string],
	// Extra args to be passed to the trufflehog cli.
	withExtraArgs Optional[string],
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
	// Image to use for the runner.
	runnerImage Optional[string],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token to use for authentication.
	token Optional[*Secret],
) *Container {
	// initializing runtime options
	opts := ActionsRuntimeRunOpts{
		Branch:      branch.GetOr(""),
		Repo:        repo.GetOr(""),
		RunnerDebug: runnerDebug.GetOr(false),
		RunnerImage: runnerImage.GetOr(""),
		Source:      source.GetOr(nil),
		Tag:         tag.GetOr(""),
		Token:       token.GetOr(nil),
	}

	return dag.ActionsRuntime().
		Run("trufflesecurity/trufflehog@main", opts).
		WithInput("path", withPath).
		WithInput("base", withBase.GetOr("")).
		WithInput("head", withHead.GetOr("")).
		WithInput("extra_args", withExtraArgs.GetOr("")).
		Sync()
}