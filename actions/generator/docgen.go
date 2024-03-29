package main

import (
	"fmt"
	"strings"
)

// generateModuleREADME generates the README.md file for the given custom action.
func generateModuleREADME(ca *CustomAction, daggerVersion string) *File {
	sb := &strings.Builder{}

	// Headers and Metadata with Prefix
	fmt.Fprintf(sb, "# Module: %s\n\n![dagger-min-version](https://img.shields.io/badge/dagger%%20version-%s-green)\n\n",
		ca.Meta.Name, daggerVersion)

	// Existing Description code
	if ca.Meta.Description != "" {
		fmt.Fprintf(sb, "%s\n\n", ca.Meta.Description)
	}

	// Clarified Notes for generation
	generatorURL := "[actions-generator](https://github.com/aweris/gale/tree/main/daggerverse/actions/generator)"
	actionURL := fmt.Sprintf("[%s](https://github.com/%s)", ca.Repo, ca.Repo)
	fmt.Fprintf(sb, "This module is automatically generated using %s. It is a Dagger-compatible adaptation of the original %s action.\n\n", generatorURL, actionURL)

	// Lead-in for Commands section
	fmt.Fprintf(sb, "## How to Use\n\nRun the following command run this action:\n\n")
	fmt.Fprintf(sb, "```shell\ndagger call -m github.com/aweris/gale/gha/%s run [flags]\n```\n\n", ca.Repo)

	// Updated Flags section
	fmt.Fprintf(sb, "## Flags\n\n")
	fmt.Fprintf(sb, "### Action Inputs\n\n")

	if len(ca.Meta.Inputs) > 0 {
		writeTableHeader(sb, []string{"Name", "Required", "Description", "Default"})

		// use sorted keys to ensure consistent output
		for _, value := range ca.Meta.Inputs {
			fmt.Fprintf(sb, "| %s | %t | %s | %s |\n", value.Name, value.Required, value.Description, value.Default)
		}
	} else {
		sb.WriteString("This action has no inputs.")
	}

	// Action Runtime Configuration
	fmt.Fprintf(sb, "\n\n### Action Runtime Inputs\n\n")
	writeTableHeader(sb, []string{"Flag", "Required", "Description"})

	type flagInfo struct {
		Required string
		Desc     string
	}

	flags := map[string]flagInfo{
		"--source": {
			Required: "Conditional",
			Desc:     "The directory containing the repository source. Either `--source` or `--repo` must be provided; `--source` takes precedence.",
		},
		"--repo": {
			Required: "Conditional",
			Desc:     "The name of the repository (owner/name). Either `--source` or `--repo` must be provided; `--source` takes precedence.",
		},
		"--tag": {
			Required: "Conditional",
			Desc:     "Tag name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence.",
		},
		"--branch": {
			Required: "Conditional",
			Desc:     "Branch name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence.",
		},
		"--container": {
			Required: "Optional",
			Desc:     "Container to use for the runner.",
		},
		"--runner-debug": {
			Required: "Optional",
			Desc:     "Enables debug mode.",
		},
		"--token": {
			Required: "Optional",
			Desc:     "GitHub token is optional for running the action. However, be aware that certain custom actions may require a token and could fail if it's not provided.",
		},
	}

	// use sorted keys to ensure consistent output
	for _, flag := range getSortedKeys(flags) {
		info := flags[flag]
		fmt.Fprintf(sb, "| %s | %s | %s |\n", flag, info.Required, info.Desc)
	}

	return dag.Directory().
		WithNewFile("README.md", sb.String()).
		File("README.md")
}

// writeTableHeader writes the table header to the given strings.Builder.
func writeTableHeader(sb *strings.Builder, headers []string) {
	sb.WriteString("| ")
	for _, header := range headers {
		sb.WriteString(header + " | ")
	}
	sb.WriteString("\n| ")
	for range headers {
		sb.WriteString("------| ")
	}
	sb.WriteString("\n")
}
