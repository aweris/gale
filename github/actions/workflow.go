package actions

import (
	"context"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"dagger.io/dagger"
)

// Workflows represents a collection of workflows.
type Workflows map[string]*Workflow

// LoadWorkflows loads the workflows from the .github/workflows directory.
func LoadWorkflows(ctx context.Context, client *dagger.Client) (Workflows, error) {
	dir := client.Host().Directory(".github/workflows")

	entries, err := dir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make(map[string]*Workflow)

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)

			workflow, err := loadWorkflow(ctx, filepath.Join(".github/workflows", entry), file)
			if err != nil {
				return nil, err
			}

			workflows[workflow.Name] = workflow
		}
	}

	return workflows, nil
}

// Workflow represents a GitHub Actions workflow.
type Workflow struct {
	// path is the relative path to the workflow file.
	path string

	// file is the workflow file itself.
	file *dagger.File

	// Name is the name of the workflow
	Name string `yaml:"name"`

	// Environment is the environment variables used in the workflow
	Environment Environment `yaml:"env"`

	// Jobs is the list of jobs in the workflow.
	Jobs Jobs `yaml:"jobs"`
}

func loadWorkflow(ctx context.Context, path string, file *dagger.File) (*Workflow, error) {
	content, err := file.Contents(ctx)
	if err != nil {
		return nil, err
	}

	var workflow Workflow

	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return nil, err
	}

	workflow.path = path
	workflow.file = file

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = path
	}

	return &workflow, nil
}
