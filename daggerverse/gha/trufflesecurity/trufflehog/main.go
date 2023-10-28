// Code generated by actions-generator. DO NOT EDIT.

package main

// Dagger module for executing the trufflesecurity/trufflehog GitHub Action.
type Trufflehog struct{}

// Runs the trufflesecurity/trufflehog GitHub Action.
func (m Trufflehog) Run(
	// Start scanning from here (usually main branch).
	withBase Optional[string],
	// Scan commits until here (usually dev branch).
	withHead Optional[string],
	// Extra args to be passed to the trufflehog cli.
	withExtraArgs Optional[string],
	// Repository path
	withPath string,
	// Directory containing the repository source. Takes precedence over `--repo`.
	source Optional[*Directory],
	// Repository name, format: owner/name. Takes precedence over `--source`.
	repo Optional[string],
	// Tag name to check out. Only works with `--repo`. Takes precedence over `--branch`.
	tag Optional[string],
	// Branch name to check out. Only works with `--repo`.
	branch Optional[string],
	// Image for the runner.
	runnerImage Optional[string],
	// Enables debug mode.
	runnerDebug Optional[bool],
	// GitHub token. May be required for certain actions.
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
		WithInput("base", withBase.GetOr("")).
		WithInput("head", withHead.GetOr("")).
		WithInput("extra_args", withExtraArgs.GetOr("")).
		WithInput("path", withPath).
		Sync()
}
