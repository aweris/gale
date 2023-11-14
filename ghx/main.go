package main

import (
	stdContext "context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/model"
	"github.com/aweris/gale/ghx/context"
)

func main() {
	stdctx := stdContext.Background()

	client, err := dagger.Connect(stdctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fmt.Printf("failed to get dagger client: %v", err)
		os.Exit(1)
	}

	// Load context
	ctx, err := context.New(stdctx, client)
	if err != nil {
		fmt.Printf("failed to load context: %v", err)
		os.Exit(1)
	}

	cfg := ctx.GhxConfig

	// Load workflow
	wf, err := LoadWorkflow(cfg, filepath.Join(cfg.HomeDir, "run", "workflow.yaml"))
	if err != nil {
		fmt.Printf("could not load workflow: %v", err)
		os.Exit(1)
	}

	// Create task runner for the workflow
	runner, err := planWorkflow(wf, cfg.Job)
	if err != nil {
		fmt.Printf("failed to plan workflow: %v", err)
		os.Exit(1)
	}

	// Run the workflow
	runner.Run(ctx)
}

func LoadWorkflow(cfg context.GhxConfig, path string) (model.Workflow, error) {
	var workflow model.Workflow

	if err := fs.ReadYAMLFile(path, &workflow); err != nil {
		return workflow, err
	}

	// set workflow path
	workflow.Path = cfg.Workflow

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = cfg.Workflow
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

	return workflow, nil
}
