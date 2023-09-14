package core

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/log"
)

type CustomAction struct {
	Path string            // Path to the custom action
	Meta *CustomActionMeta // Meta information about the custom action
	Dir  *dagger.Directory // Dir where the custom action is located
}

// CustomActionMeta represents a metadata for a GitHub Action. It contains all the information needed to run the action.
// The metadata is loaded from the action.yml | action.yaml file in the action repository.
//
// See more details at https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions
type CustomActionMeta struct {
	Name        string                        `yaml:"name"`        // Name is the name of the action.
	Author      string                        `yaml:"author"`      // Author is the author of the action.
	Description string                        `yaml:"description"` // Description is the description of the action.
	Inputs      map[string]CustomActionInput  `yaml:"inputs"`      // Inputs is a map of input names to their definitions.
	Outputs     map[string]CustomActionOutput `yaml:"outputs"`     // Outputs is a map of output names to their definitions.
	Runs        CustomActionRuns              `yaml:"runs"`        // Runs is the definition of how the action is run.
	Branding    Branding                      `yaml:"branding"`    // Branding is the branding information for the action.
}

// CustomActionInput represents an input for a GitHub Action.
type CustomActionInput struct {
	Description        string `yaml:"description"`        // Description is the description of the input.
	Default            string `yaml:"default"`            // Default is the default value of the input.
	Required           bool   `yaml:"required"`           // Required is whether the input is required.
	DeprecationMessage string `yaml:"deprecationMessage"` // DeprecationMessage is the message to display when the input is used.
}

// CustomActionOutput represents an output for a GitHub Action.
type CustomActionOutput struct {
	Description string `yaml:"description"` // Description is the description of the output.
	Value       string `yaml:"value"`       // Value is the value of the output.
}

// CustomActionRuns represents the definition of how a GitHub Action is run.
type CustomActionRuns struct {
	Using          CustomActionRunsUsing `yaml:"using"`           // Using is the method used to run the action.
	Env            map[string]string     `yaml:"env"`             // Env is the environment variables used to run the action.
	Main           string                `yaml:"main"`            // Main is the path to the main entrypoint for the action. This is only used by javascript actions.
	Pre            string                `yaml:"pre"`             // Pre is the path to the pre entrypoint for the action. This is only used by javascript actions.
	PreIf          string                `yaml:"pre-if"`          // PreIf is the condition for running the pre entrypoint. This is only used by javascript actions.
	Post           string                `yaml:"post"`            // Post is the path to the post entrypoint for the action. This is only used by javascript actions.
	PostIf         string                `yaml:"post-if"`         // PostIf is the condition for running the post entrypoint. This is only used by javascript actions.
	Steps          []Step                `yaml:"steps"`           // Steps is the list of steps to run for the action. This is only used by composite actions.
	Image          string                `yaml:"image"`           // Image is the image used to run the action. This is only used by docker actions.
	PreEntrypoint  string                `yaml:"pre-entrypoint"`  // PreEntrypoint is the pre-entrypoint used to run the action. This is only used by docker actions.
	Entrypoint     string                `yaml:"entrypoint"`      // Entrypoint is the entrypoint used to run the action. This is only used by docker actions.
	PostEntrypoint string                `yaml:"post-entrypoint"` // PostEntrypoint is the post-entrypoint used to run the action. This is only used by docker actions.
	Args           []string              `yaml:"args"`            // Args is the arguments used to run the action. This is only used by docker actions.
}

// CustomActionRunsUsing represents the method used to run a GitHub Action.
type CustomActionRunsUsing string

var (
	// ActionRunsUsingComposite is the value for ActionRunsUsing when the action is a composite action.
	ActionRunsUsingComposite CustomActionRunsUsing = "composite"

	// ActionRunsUsingDocker is the value for ActionRunsUsing when the action is a docker action.
	ActionRunsUsingDocker CustomActionRunsUsing = "docker"

	// ActionRunsUsingNode20 is the value for ActionRunsUsing when the action is a javascript action using node 20.
	ActionRunsUsingNode20 CustomActionRunsUsing = "node20"

	// ActionRunsUsingNode16 is the value for ActionRunsUsing when the action is a javascript action using node 16.
	ActionRunsUsingNode16 CustomActionRunsUsing = "node16"

	// ActionRunsUsingNode12 is the value for ActionRunsUsing when the action is a javascript action using node 12.
	// Deprecated by GitHub. will be removed in the future. Added for backwards compatibility.
	ActionRunsUsingNode12 CustomActionRunsUsing = "node12"
)

// UnmarshalYAML unmarshal the action runs using value and validate it against the supported values.
func (a *CustomActionRunsUsing) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val string

	if err := unmarshal(&val); err != nil {
		return err
	}

	// Force input to lowercase for case-insensitive comparison
	using := CustomActionRunsUsing(strings.ToLower(val))

	// unmarshal all unsupported values as invalid
	switch using {
	case ActionRunsUsingComposite, ActionRunsUsingDocker, ActionRunsUsingNode20, ActionRunsUsingNode16, ActionRunsUsingNode12:
		*a = using
	default:
		return fmt.Errorf("invalid value for using: %s", using)
	}

	return nil
}

// Branding represents the branding information for a GitHub Action.
type Branding struct {
	Color string `yaml:"color"` // Color is the color of the action.
	Icon  string `yaml:"icon"`  // Icon is the icon of the action.
}

// PreCondition returns if the action has a pre-condition and the pre-if value for supported actions. If action
// does not have a pre step or the pre step is not supported, the method returns false for the first return value.
func (c *CustomActionRuns) PreCondition() (bool, string) {
	var pre string

	switch c.Using {
	case ActionRunsUsingDocker:
		pre = c.PreEntrypoint
	case ActionRunsUsingNode16, ActionRunsUsingNode12:
		pre = c.Pre
	default:
		pre = "" // all other types of actions do not have a pre-condition
	}

	// if pre is not set, return false
	if pre == "" {
		return false, ""
	}

	return true, c.PreIf
}

// PostCondition returns if the action has a post-condition and the post-if value for supported actions. If action
// does not have a post step or the post step is not supported, the method returns false for the first return value.
func (c *CustomActionRuns) PostCondition() (bool, string) {
	var post string

	switch c.Using {
	case ActionRunsUsingDocker:
		post = c.PostEntrypoint
	case ActionRunsUsingNode16, ActionRunsUsingNode12:
		post = c.Post
	default:
		post = "" // all other types of actions do not have a post-condition
	}

	// if post is not set, return false
	if post == "" {
		return false, ""
	}

	return true, c.PostIf
}

// LoadActionFromSource loads an action from given source to the target directory. If the source is a local action,
// the target directory will be the same as the source. If the source is a remote action, the action will be downloaded
// to the target directory using the source as the reference(e.g. {target}/{owner}/{repo}/{path}@{ref}).
func LoadActionFromSource(ctx context.Context, client *dagger.Client, source, targetDir string) (*CustomAction, error) {
	var target string

	// no need to load action if it is a local action
	if isLocalAction(source) {
		target = source
	} else {
		target = filepath.Join(targetDir, source)

		// ensure action exists locally
		if err := ensureActionExistsLocally(ctx, client, source, target); err != nil {
			return nil, err
		}
	}

	dir, err := getActionDirectory(client, target)
	if err != nil {
		return nil, err
	}

	meta, err := getCustomActionMeta(ctx, dir)
	if err != nil {
		return nil, err
	}

	return &CustomAction{Meta: meta, Path: target, Dir: dir}, nil
}

// isLocalAction checks if the given source is a local action
func isLocalAction(source string) bool {
	return strings.HasPrefix(source, "./") || filepath.IsAbs(source) || strings.HasPrefix(source, "/")
}

// ensureActionExistsLocally ensures that the action exists locally. If the action does not exist locally, it will be
// downloaded from the source to the target directory.
func ensureActionExistsLocally(ctx context.Context, client *dagger.Client, source, target string) error {
	// check if action exists locally
	exist, err := fs.Exists(target)
	if err != nil {
		return fmt.Errorf("failed to check if action exists locally: %w", err)
	}

	// do nothing if target path already exists
	if exist {
		log.Debugf("action already exists locally", "source", source, "target", target)
		return nil
	}

	log.Debugf("action does not exist locally, downloading...", "source", source, "target", target)

	dir, err := getActionDirectory(client, source)
	if err != nil {
		return err
	}

	// export the action to the target directory
	_, err = dir.Export(ctx, target)
	if err != nil {
		return err
	}

	return nil
}

// getCustomActionMeta returns the meta information about the custom action from the action directory.
func getCustomActionMeta(ctx context.Context, actionDir *dagger.Directory) (*CustomActionMeta, error) {
	var meta CustomActionMeta

	file, err := findActionMetadataFileName(ctx, actionDir)
	if err != nil {
		return nil, err
	}

	content, err := actionDir.File(file).Contents(ctx)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(content), &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// getActionDirectory returns the directory of the action from given source.
func getActionDirectory(client *dagger.Client, source string) (*dagger.Directory, error) {
	// if path is relative, use the host to resolve the path
	if isLocalAction(source) {
		return client.Host().Directory(source), nil
	}

	// if path is not a relative path, it must be a remote repository in the format "{owner}/{repo}/{path}@{ref}"
	// if {path} is not present in the input string, an empty string is returned for the path component.
	actionRepo, actionPath, actionRef, err := parseRepoRef(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo ref %s: %v", source, err)
	}

	// TODO: handle enterprise github instances as well
	gitRepo := client.Git(path.Join("github.com", actionRepo))

	var gitRef *dagger.GitRef

	switch DetermineRefTypeFromRepo(actionRepo, actionRef) {
	case RefTypeBranch:
		gitRef = gitRepo.Branch(actionRef)
	case RefTypeTag:
		gitRef = gitRepo.Tag(actionRef)
	case RefTypeCommit:
		gitRef = gitRepo.Commit(actionRef)
	default:
		return nil, fmt.Errorf("failed to determine ref type for %s: %v", source, err)
	}

	// if path is empty, use the root of the repo as the action directory
	if actionPath == "" {
		actionPath = "."
	}

	return gitRef.Tree().Directory(actionPath), nil
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
