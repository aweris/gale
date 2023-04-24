package run

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/executor"
	"github.com/aweris/gale/gha"
)

// NewCommand creates a new run command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a workflow",
		Long:  `Run a Github Actions workflow locally.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkflow()
		},
	}

	return cmd
}

func runWorkflow() error {
	// Create a context to pass to Dagger.
	ctx := context.Background()

	// Connect to Dagger
	client, clientErr := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if clientErr != nil {
		return clientErr
	}

	// Load the workflows from the .github/workflows directory.
	workflows, loadErr := gha.LoadWorkflows(ctx, client)
	if loadErr != nil {
		return loadErr
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
		return execErr
	}

	return nil
}
