package model

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"dagger.io/dagger"
)

// ActionStage represents the stage of an action. It can be pre, main or post.
type ActionStage string

const (
	ActionStagePre  ActionStage = "pre"
	ActionStageMain ActionStage = "main"
	ActionStagePost ActionStage = "post"
)

// Action represents a metadata for a GitHub Action. It contains all the information needed to run the action.
// The metadata is loaded from the action.yml | action.yaml file in the action repository.
//
// See more details at https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type Action struct {
	// Name is the name of the action.
	Name string `yaml:"name"`

	// Author is the author of the action.
	Author string `yaml:"author"`

	// Description is the description of the action.
	Description string `yaml:"description"`

	// Inputs is a map of input names to their definitions.
	Inputs map[string]ActionInput `yaml:"inputs"`

	// Outputs is a map of output names to their definitions.
	Outputs map[string]ActionInput `yaml:"outputs"`

	// Runs is the definition of how the action is run.
	Runs ActionRuns `yaml:"runs"`

	// Branding is the branding information for the action.
	Branding Branding `yaml:"branding"`

	// Directory is the directory where source files for the action are located.
	Directory *dagger.Directory `yaml:"-"`
}

// ActionInput represents an input for a GitHub Action.
type ActionInput struct {
	// Description is the description of the input.
	Description string `yaml:"description"`

	// Default is the default value of the input.
	Default string `yaml:"default"`

	// Required is whether the input is required.
	Required bool `yaml:"required"`

	// DeprecationMessage is the message to display when the input is used.
	DeprecationMessage string `yaml:"deprecationMessage"`
}

// ActionOutput represents an output for a GitHub Action.
type ActionOutput struct {
	// Description is the description of the output.
	Description string `yaml:"description"`

	// Value is the value of the output.
	Value string `yaml:"value"`
}

// ActionRuns represents the definition of how a GitHub Action is run.
type ActionRuns struct {
	// Using is the method used to run the action.
	Using ActionRunsUsing `yaml:"using"`

	// Env is a map of environment variables to their values.
	Env map[string]string `yaml:"env"`

	// Main is the path to the main entrypoint for the action. This is only used by javascript actions.
	Main string `yaml:"main"`

	// Pre is the path to the pre entrypoint for the action. This is only used by javascript actions.
	Pre string `yaml:"pre"`

	// PreIf is the condition for running the pre entrypoint. This is only used by javascript actions.
	PreIf string `yaml:"pre-if"`

	// Post is the path to the post entrypoint for the action. This is only used by javascript actions.
	Post string `yaml:"post"`

	// PostIf is the condition for running the post entrypoint. This is only used by javascript actions.
	PostIf string `yaml:"post-if"`

	// Steps is the list of steps to run for the action. This is only used by composite actions.
	Steps []Step `yaml:"steps"`

	// Image is the image used to run the action. This is only used by docker actions.
	Image string `yaml:"image"`

	// PreEntrypoint is the pre-entrypoint used to run the action. This is only used by docker actions.
	PreEntrypoint string `yaml:"pre-entrypoint"`

	// Entrypoint is the entrypoint used to run the action. This is only used by docker actions.
	Entrypoint string `yaml:"entrypoint"`

	// PostEntrypoint is the post-entrypoint used to run the action. This is only used by docker actions.
	PostEntrypoint string `yaml:"post-entrypoint"`

	// Args is the arguments used to run the action. This is only used by docker actions.
	Args []string `yaml:"args"`
}

// ActionRunsUsing represents the method used to run a GitHub Action.
type ActionRunsUsing string

var (
	// ActionRunsUsingComposite is the value for ActionRunsUsing when the action is a composite action.
	ActionRunsUsingComposite ActionRunsUsing = "composite"

	// ActionRunsUsingDocker is the value for ActionRunsUsing when the action is a docker action.
	ActionRunsUsingDocker ActionRunsUsing = "docker"

	// ActionRunsUsingNode16 is the value for ActionRunsUsing when the action is a javascript action using node 16.
	ActionRunsUsingNode16 ActionRunsUsing = "node16"

	// ActionRunsUsingNode12 is the value for ActionRunsUsing when the action is a javascript action using node 12.
	// Deprecated by GitHub. will be removed in the future. Added for backwards compatibility.
	ActionRunsUsingNode12 ActionRunsUsing = "node12"
)

// UnmarshalYAML unmarshal the action runs using value and validate it against the supported values.
func (a *ActionRunsUsing) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val string

	if err := unmarshal(&val); err != nil {
		return err
	}

	// Force input to lowercase for case-insensitive comparison
	using := ActionRunsUsing(strings.ToLower(val))

	// unmarshal all unsupported values as invalid
	switch using {
	case ActionRunsUsingComposite, ActionRunsUsingDocker, ActionRunsUsingNode16, ActionRunsUsingNode12:
		*a = using
	default:
		return fmt.Errorf("invalid value for using: %s", using)
	}

	return nil
}

// Branding represents the branding information for a GitHub Action.
type Branding struct {
	// Color is the color of the action.
	Color string `yaml:"color"`

	// Icon is the icon of the action.
	Icon string `yaml:"icon"`
}

// LoadActionFromSource loads an action from given source. Source can be a local directory or a remote repository.
func LoadActionFromSource(ctx context.Context, client *dagger.Client, src string) (*Action, error) {
	var dir *dagger.Directory

	dir, dirErr := getActionDirectory(client, src)
	if dirErr != nil {
		return nil, dirErr
	}

	file, findActionErr := findActionMetadataFileName(ctx, dir)
	if findActionErr != nil {
		return nil, findActionErr
	}

	content, contentErr := dir.File(file).Contents(ctx)
	if contentErr != nil {
		return nil, fmt.Errorf("failed to read %s/%s: %v", src, file, contentErr)
	}

	var action Action

	if unmarshalErr := yaml.Unmarshal([]byte(content), &action); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to unmarshal %s/%s: %v", src, file, unmarshalErr)
	}

	action.Directory = dir

	return &action, nil
}

// getActionDirectory returns the directory of the action from given source.
func getActionDirectory(client *dagger.Client, src string) (*dagger.Directory, error) {
	// if path is relative, use the host to resolve the path
	if strings.HasPrefix(src, "./") || filepath.IsAbs(src) || strings.HasPrefix(src, "/") {
		return client.Host().Directory(src), nil
	}

	// if path is not a relative path, it must be a remote repository in the format "{owner}/{repo}/{path}@{ref}"
	// if {path} is not present in the input string, an empty string is returned for the path component.

	actionRepo, actionPath, actionRef, err := parseRepoRef(src)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo ref %s: %v", src, err)
	}

	// if path is empty, use the root of the repo as the action directory
	if actionPath == "" {
		actionPath = "."
	}

	// TODO: handle enterprise github instances as well
	// TODO: handle ref type (branch, tag, commit) currently only tags are supported
	return client.Git(path.Join("github.com", actionRepo)).Tag(actionRef).Tree().Directory(actionPath), nil
}

// findActionMetadataFileName finds the action.yml or action.yaml file in the root of the action directory.
func findActionMetadataFileName(ctx context.Context, dir *dagger.Directory) (string, error) {
	// list all entries in the root of the action directory
	entries, entriesErr := dir.Entries(ctx)
	if entriesErr != nil {
		return "", fmt.Errorf("failed to list entries for: %v", entriesErr)
	}

	file := ""

	// find action.yml or action.yaml exists in the root of the action repo
	for _, entry := range entries {
		if entry == "action.yml" || entry == "action.yaml" {
			file = entry
			break
		}
	}

	// if action.yml or action.yaml does not exist, return an error
	if file == "" {
		return "", fmt.Errorf("action.yml or action.yaml not found in the root of the action directory")
	}

	return file, nil
}

// parseRepoRef parses a string in the format "{owner}/{repo}/{path}@{ref}" and returns the parsed components.
// If {path} is not present in the input string, an empty string is returned for the path component.
func parseRepoRef(input string) (repo string, path string, ref string, err error) {
	regex := regexp.MustCompile(`^([^/]+)/([^/@]+)(?:/([^@]+))?@(.+)$`)
	matches := regex.FindStringSubmatch(input)

	if len(matches) == 0 {
		err = fmt.Errorf("invalid input format: %q", input)
		return
	}

	repo = strings.Join([]string{matches[1], matches[2]}, "/")
	path = matches[3]
	ref = matches[4]

	return
}
