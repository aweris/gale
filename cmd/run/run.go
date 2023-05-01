package run

import (
	"context"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/journal"
	"github.com/aweris/gale/logger"
	runnerpkg "github.com/aweris/gale/runner"
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

	journalW, journalR := journal.Pipe()

	log := logger.NewLogger(logger.WithVerbose(false), logger.WithJournalR(journalR))

	// Connect to Dagger
	client, clientErr := dagger.Connect(ctx, dagger.WithLogOutput(journalW))
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()

	// Load the workflows from the .github/workflows directory.
	workflows, loadErr := gha.LoadWorkflows(ctx, client)
	if loadErr != nil {
		return loadErr
	}

	// Pick a workflow and job to run manually to test.
	workflow := workflows["Clone"]
	job := workflow.Jobs["clone"]

	// Create runner
	runner, err := runnerpkg.NewRunner(ctx, client, log, gha.NewDummyContext(), workflow, job)
	if err != nil {
		return err
	}

	runner.Run(ctx)

	return nil
}
