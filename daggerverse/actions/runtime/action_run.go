package main

import (
	"fmt"

	"actions-runtime/model"
)

type ActionRun struct {
	Config ActionRunConfig
}

// ActionRunConfig holds the configuration of a action run.
type ActionRunConfig struct {
	// Directory containing the repository source.
	Source *Directory

	// Name of the repository. Format: owner/name.
	Repo string

	// Branch name to check out. Only one of branch or tag can be used. Precedence: tag, branch.
	Branch string

	// Tag name to check out. Only one of branch or tag can be used. Precedence: tag, branch.
	Tag string

	// The action to run. it should be in the format of <action>@<version>
	Uses string

	// Environment variables for the action. Format: name=value.
	Env []string

	// Input parameters for the action. Format: name=value.
	With []string

	// Name of the event that triggered the workflow. e.g. push
	Event string

	// File with the complete webhook event payload.
	EventFile *File

	// Image to use for the runner. If --image and --container provided together, --image takes precedence.
	Image string

	// Container to use for the runner. If --image and --container provided together, --image takes precedence.
	Container *Container

	// Enables debug mode.
	RunnerDebug bool

	// GitHub token to use for authentication.
	Token *Secret
}

func (ar *ActionRun) WithEnv(name string, value string) *ActionRun {
	ar.Config.Env = append(ar.Config.Env, fmt.Sprintf("%s=%s", name, value))
	return ar
}

func (ar *ActionRun) WithInput(name string, value string) *ActionRun {
	ar.Config.With = append(ar.Config.With, fmt.Sprintf("%s=%s", name, value))
	return ar
}

func (ar *ActionRun) Sync() (*Container, error) {
	wf, err := getWorkflowFile(ar.Config.Uses, ar.Config.Env, ar.Config.With)
	if err != nil {
		return nil, err
	}

	opts := GaleRunOpts{
		Source:       ar.Config.Source,
		Repo:         ar.Config.Repo,
		Tag:          ar.Config.Tag,
		Branch:       ar.Config.Branch,
		WorkflowFile: wf,
		Event:        ar.Config.Event,
		EventFile:    ar.Config.EventFile,
		Image:        ar.Config.Image,
		Container:    ar.Config.Container,
		RunnerDebug:  ar.Config.RunnerDebug,
		Token:        ar.Config.Token,
	}

	return dag.Gale().Run(opts).Sync(), nil
}

// getWorkflowFile returns a workflow file with the given uses, env, and with.
func getWorkflowFile(uses string, env, with []string) (*File, error) {
	w := &model.Workflow{
		Jobs: map[string]model.Job{
			"dagger": {
				Steps: []model.Step{
					{
						ID:          "dagger",
						Uses:        uses,
						Environment: parseKeyValues(env),
						With:        parseKeyValues(with),
					},
				},
			},
		},
	}

	return marshalWorkflowToFile(w)
}
