package core

import (
	"fmt"
	"strings"

	"dagger.io/dagger"
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
	case ActionRunsUsingComposite, ActionRunsUsingDocker, ActionRunsUsingNode16, ActionRunsUsingNode12:
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
