package main

import (
	stdContext "context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/common/fs"
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
	workflows, err := LoadWorkflows(cfg.WorkflowsDir)
	if err != nil {
		fmt.Printf("failed to load workflows: %v", err)
		os.Exit(1)
	}

	wf, ok := workflows[cfg.Workflow]
	if !ok {
		fmt.Printf("workflow %s not found", cfg.Workflow)
		os.Exit(1)
	}

	// Create task runner for the workflow
	runner, err := planWorkflow(wf, cfg.Job)
	if err != nil {
		fmt.Printf("failed to plan workflow: %v", err)
		os.Exit(1)
	}

	// Run the workflow
	result, _ := runner.Run(ctx)

	err = fs.WriteJSONFile("/home/runner/_temp/ghx/result.json", &result)
	if err != nil {
		fmt.Printf("failed to write result: %v", err)
		os.Exit(1)
	}
}
