package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/model"
)

func loadCustomAction(source string) (*CustomAction, error) {
	regex := regexp.MustCompile(`^([^/]+)/([^/@]+)(?:/([^@]+))?@(.+)$`)
	matches := regex.FindStringSubmatch(source)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid input format: %q", source)
	}

	var (
		owner = matches[1]
		name  = matches[2]
		repo  = strings.Join([]string{matches[1], matches[2]}, "/")
		ref   = matches[4]
	)

	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.yml", repo, ref))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading action.yml: %v\n", err)
	}
	defer resp.Body.Close()

	// if action.yml does not exist, try action.yaml instead. // FIXME: Find a better way to do this.
	if resp.StatusCode == http.StatusNotFound {
		println(fmt.Sprintf("==> action.yml not found for %s@%s. Trying action.yaml instead...", repo, ref))

		resp, err = http.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.yaml", repo, ref))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error downloading action.yaml: %v\n", err)
		}
		// defer already called above, so no need to call again
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading action.yaml: %v", resp.Status)
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading action.yml: %v", err)
	}

	var meta model.CustomActionMeta

	if err := yaml.Unmarshal(contents, &meta); err != nil {
		return nil, fmt.Errorf("error unmarshalling action.yml: %v", err)
	}

	return &CustomAction{
		Owner:    owner,
		RepoName: name,
		Repo:     repo,
		Ref:      ref,
		Meta:     toCustomActionMeta(&meta),
	}, nil
}

func toCustomActionMeta(meta *model.CustomActionMeta) CustomActionMeta {
	return CustomActionMeta{
		Name:        meta.Name,
		Author:      meta.Author,
		Description: meta.Description,
		Inputs:      toCustomActionInputs(meta.Inputs),
		Outputs:     toCustomActionOutputs(meta.Outputs),
		Runs: CustomActionRuns{
			Using:          CustomActionRunsUsing(meta.Runs.Using),
			Env:            ConvertMapToKVSlice(meta.Runs.Env),
			Main:           meta.Runs.Main,
			Pre:            meta.Runs.Pre,
			PreIf:          meta.Runs.PreIf,
			Post:           meta.Runs.Post,
			PostIf:         meta.Runs.PostIf,
			Steps:          toCustomActionSteps(meta.Runs.Steps),
			Image:          meta.Runs.Image,
			PreEntrypoint:  meta.Runs.PreEntrypoint,
			Entrypoint:     meta.Runs.Entrypoint,
			PostEntrypoint: meta.Runs.PostEntrypoint,
			Args:           meta.Runs.Args,
		},
		Branding: Branding{
			Icon:  meta.Branding.Icon,
			Color: meta.Branding.Color,
		},
	}
}

func toCustomActionInputs(inputs map[string]model.CustomActionInput) []CustomActionInput {
	// ensure keys are sorted to ensure consistent output
	keys := getSortedKeys(inputs)

	var result []CustomActionInput

	// use sorted keys to convert map to slice
	for _, k := range keys {
		v := inputs[k]

		result = append(result, CustomActionInput{
			Name:               k,
			Description:        v.Description,
			Default:            v.Default,
			Required:           v.Required,
			DeprecationMessage: v.DeprecationMessage,
		})
	}

	return result
}

func toCustomActionOutputs(outputs map[string]model.CustomActionOutput) []CustomActionOutput {
	// ensure keys are sorted to ensure consistent output
	keys := getSortedKeys(outputs)

	var result []CustomActionOutput

	// use sorted keys to convert map to slice
	for _, k := range keys {
		v := outputs[k]

		result = append(result, CustomActionOutput{
			Name:        k,
			Description: v.Description,
			Value:       v.Value,
		})
	}

	return result
}

func toCustomActionSteps(steps []model.Step) []Step {
	var result []Step

	for _, v := range steps {
		result = append(result, Step{
			Name:            v.Name,
			Uses:            v.Uses,
			With:            ConvertMapToKVSlice(v.With),
			Environment:     ConvertMapToKVSlice(v.Environment),
			Run:             v.Run,
			Shell:           v.Shell,
			ContinueOnError: v.ContinueOnError,
			TimeoutMinutes:  v.TimeoutMinutes,
		})
	}

	return result
}

type CustomAction struct {
	Repo     string           // Repo is the repository name in the form "owner/repo".
	RepoName string           // RepoName is the repository name in the form "repo".
	Owner    string           // Owner is the repository owner.
	Ref      string           // Ref is the git ref of the custom action.
	Meta     CustomActionMeta // Meta contains action.yml contents for the custom action.
}

type CustomActionMeta struct {
	Name        string               // Name is the name of the custom action.
	Author      string               // Author is the author of the custom action.
	Description string               // Description is the description of the custom action.
	Inputs      []CustomActionInput  // Inputs is a map of input names to their definitions.
	Outputs     []CustomActionOutput // Outputs is a map of output names to their definitions.
	Runs        CustomActionRuns     // Runs is the definition of how the action is run.
	Branding    Branding             // Branding is the branding information for the action.

}

// CustomActionInput represents an input for a GitHub Action.
type CustomActionInput struct {
	Name               string // Name is the name of the input.
	Description        string // Description is the description of the input.
	Default            string // Default is the default value of the input.
	Required           bool   // Required is whether the input is required.
	DeprecationMessage string // DeprecationMessage is the message to display when the input is used.
}

// CustomActionOutput represents an output for a GitHub Action.
type CustomActionOutput struct {
	Name        string // Name is the name of the output.
	Description string // Description is the description of the output.
	Value       string // Value is the value of the output.
}

// CustomActionRuns represents the definition of how a GitHub Action is run.
type CustomActionRuns struct {
	Using          CustomActionRunsUsing // Using is the method used to run the action.
	Env            []KV                  // Env is the environment variables used to run the action.
	Main           string                // Main is the path to the main entrypoint for the action. This is only used by javascript actions.
	Pre            string                // Pre is the path to the pre entrypoint for the action. This is only used by javascript actions.
	PreIf          string                // PreIf is the condition for running the pre entrypoint. This is only used by javascript actions.
	Post           string                // Post is the path to the post entrypoint for the action. This is only used by javascript actions.
	PostIf         string                // PostIf is the condition for running the post entrypoint. This is only used by javascript actions.
	Steps          []Step                // Steps is the list of steps to run for the action. This is only used by composite actions.
	Image          string                // Image is the image used to run the action. This is only used by docker actions.
	PreEntrypoint  string                // PreEntrypoint is the pre-entrypoint used to run the action. This is only used by docker actions.
	Entrypoint     string                // Entrypoint is the entrypoint used to run the action. This is only used by docker actions.
	PostEntrypoint string                // PostEntrypoint is the post-entrypoint used to run the action. This is only used by docker actions.
	Args           []string              // Args is the arguments used to run the action. This is only used by docker actions.
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

// Step represents a single task in a job context at GitHub Actions workflow
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	ID               string // ID is the unique identifier of the step.
	If               string // If is the conditional expression to run the step.
	Name             string // Name is the name of the step.
	Uses             string // Uses is the action to run for the step.
	Environment      []KV   // Environment maps environment variable names to their values.
	With             []KV   // With maps input names to their values for the step.
	Run              string // Run is the command to run for the step.
	Shell            string // Shell is the shell to use for the step.
	WorkingDirectory string // WorkingDirectory is the working directory for the step.
	ContinueOnError  bool   // ContinueOnError is a flag to continue on error.
	TimeoutMinutes   int    // TimeoutMinutes is the maximum number of minutes to run the step.
}

// Branding represents the branding information for a GitHub Action.
type Branding struct {
	Color string `yaml:"color"` // Color is the color of the action.
	Icon  string `yaml:"icon"`  // Icon is the icon of the action.
}
