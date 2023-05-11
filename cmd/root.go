package cmd

import (
	"context"
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/repository"
	runnerpkg "github.com/aweris/gale/runner"
	"github.com/aweris/gale/runner/state"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use: "gale",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
			if err != nil {
				return err
			}

			repo, err := repository.Current(ctx, client)
			if err != nil {
				return err
			}

			runner, err := runnerpkg.NewRunner(&state.BaseState{Repo: repo, Client: client})
			if err != nil {
				return err
			}

			return runner.RunWorkflow(ctx, ".github/workflows/clone.yaml")
		},
	}
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
