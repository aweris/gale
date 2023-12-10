package main

import (
	stdContext "context"
	"fmt"
	"os"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/fs"

	"ghx/context"
	"github.com/aweris/gale/common/model"
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

	ctx.Execution.Workflow = &wf

	jm, ok := wf.Jobs[ctx.GhxConfig.Job]
	if !ok {
		fmt.Printf("job %s not found", ctx.GhxConfig.Job)
		os.Exit(1)
	}

	runners, err := planJob(jm)
	if err != nil {
		fmt.Printf("failed to plan job: %v", err)
		os.Exit(1)
	}

	// FIXME: ignoring fail-fast for now. it is always true for now. Fix this later.
	// FIXME: run all runners sequentially for now. Ignoring parallelism. Fix this later.

	for _, runner := range runners {
		_, err := runner.Run(ctx)
		if err != nil {
			fmt.Printf("failed to run job: %v", err)
			os.Exit(1)
		}

	}
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
			step.Index = fmt.Sprintf("%d", ids)

			if step.ID == "" {
				step.ID = step.Index
			}

			job.Steps[ids] = step
		}

		workflow.Jobs[idj] = job
	}

	return workflow, nil
}
