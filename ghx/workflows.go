package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/ghx/core"
)

func LoadWorkflows(path string) (map[string]core.Workflow, error) {
	workflows := make(map[string]core.Workflow)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
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
	if err != nil {
		return nil, err
	}

	return workflows, nil
}
