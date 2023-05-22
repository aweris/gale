package repository

import (
	"context"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/model"
)

// Repo represents a GitHub repository. This is a wrapper around model.Repository with an additional
// field Path which is the path to the repository.
type Repo struct {
	*model.Repository

	Path      string          // path to the repository
	DataHome  string          // path to the repository data home
	Workflows model.Workflows // workflows in the repository
}

// DataPath constructs a path in the repository data home.
func (r *Repo) DataPath(path ...string) string {
	return filepath.Join(r.DataHome, filepath.Join(path...))
}

func loadWorkflows(ctx context.Context, client *dagger.Client, path string) (model.Workflows, error) {
	dir := client.Host().Directory(filepath.Join(path, ".github/workflows"))

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
	workflow.File = file

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = path
	}

	return &workflow, nil
}
