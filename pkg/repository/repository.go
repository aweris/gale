package repository

import (
	"context"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/pkg/model"
)

func LoadWorkflows(ctx context.Context, client *dagger.Client) (model.Workflows, error) {
	dir := client.Host().Directory(filepath.Join(".", ".github/workflows"))

	entries, err := dir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make(map[string]*model.Workflow)

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

func loadWorkflow(ctx context.Context, path string, file *dagger.File) (*model.Workflow, error) {
	content, err := file.Contents(ctx)
	if err != nil {
		return nil, err
	}

	var workflow model.Workflow

	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return nil, err
	}

	workflow.Path = path

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = path
	}

	return &workflow, nil
}
