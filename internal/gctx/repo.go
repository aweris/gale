package gctx

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
)

type RepoContext struct {
	Ref          core.RepositoryGitRef `json:"ref"`
	WorkflowsDir string                `json:"workflows_dir" env:"GALE_WORKFLOWS_DIR" envDefault:".github/workflows"`
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo() error {
	// load repo context from env
	rc, err := NewContextFromEnv[RepoContext]()
	if err != nil {
		return err
	}

	rc.Ref, err = core.GetRepositoryRefFromDir(".")
	if err != nil {
		return err
	}

	c.Github.setRepo(rc.Ref)

	c.Repo = rc

	return nil
}

func (c *Context) LoadWorkflows() (map[string]core.Workflow, error) {
	workflows := make(map[string]core.Workflow)

	filepath.Walk(".github/workflows", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
			var workflow core.Workflow

			if err := fs.ReadYAMLFile(path, &workflow); err != nil {
				return err
			}

			// set workflow path
			workflow.Path = path

			// if the workflow name is not provided, use the relative path to the workflow file.
			if workflow.Name == "" {
				workflow.Name = path
			}

			// update job ID and names
			for idj, job := range workflow.Jobs {
				job.ID = idj

				if job.Name == "" {
					job.Name = idj
				}

				// update step IDs if not provided
				for ids, step := range job.Steps {
					if step.ID == "" {
						step.ID = fmt.Sprintf("%d", ids)
					}

					job.Steps[ids] = step
				}

				workflow.Jobs[idj] = job
			}

			workflows[workflow.Name] = workflow
		}

		return nil
	})

	return workflows, nil
}
