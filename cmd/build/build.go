package build

import (
	"context"
	"dagger.io/dagger"
	"os"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/builder"
	"github.com/aweris/gale/github/cli"
	"github.com/aweris/gale/journal"
	"github.com/aweris/gale/logger"
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

	_, journalR := journal.Pipe()

	_ = logger.NewLogger(logger.WithJournalR(journalR))

	// Connect to Dagger
	client, clientErr := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if clientErr != nil {
		return clientErr
	}
	defer client.Close()

	repo, err := cli.CurrentRepository(ctx)
	if err != nil {
		return err
	}

	_, err = builder.NewBuilder(client, repo).Build(ctx)
	if err != nil {
		return err
	}

	return nil
}
