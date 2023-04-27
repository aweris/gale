package build

import (
	"context"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/journal"
	"github.com/aweris/gale/logger"
	"github.com/aweris/gale/runner"
)

// NewCommand creates a new run command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build a Runner image",
		Long:  `Build a Runner image`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return build()
		},
	}

	return cmd
}

func build() error {
	// Create a context to pass to Dagger.
	ctx := context.Background()

	journalW, journalR := journal.Pipe()

	_ = logger.NewLogger(logger.WithJournalR(journalR))

	// Connect to Dagger
	client, clientErr := dagger.Connect(ctx, dagger.WithLogOutput(journalW))
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()

	_, err := runner.NewBuilder(client).Build(ctx)
	if err != nil {
		return err
	}

	return nil
}
