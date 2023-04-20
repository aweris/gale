package main

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/executor"
	"github.com/aweris/gale/gha"
)

func main() {
	// Create a context to pass to Dagger.
	ctx := context.Background()

	// Connect to Dagger
	client, clientErr := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if clientErr != nil {
		fmt.Printf("Error connecting to Dagger: %v", clientErr)
		os.Exit(1)
	}

	// Load the workflows from the .github/workflows directory.
	workflows, loadErr := gha.LoadWorkflows(ctx, client)
	if loadErr != nil {
		panic(loadErr)
	}

	// Pick a workflow and job to run manually to test.
	workflow := workflows["Clone"]
	job := workflow.Jobs["clone"]

	// Create a job executor and run the job.
	je, jeErr := executor.NewJobExecutor(ctx, client, workflow, job, gha.NewDummyContext())
	if jeErr != nil {
		panic(jeErr)
	}

	execErr := je.Execute(ctx)
	if execErr != nil {
		panic(execErr)
	}
}
