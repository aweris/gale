package main

import (
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/cmd/gale/list"
	"github.com/aweris/gale/cmd/gale/run"
	"github.com/aweris/gale/cmd/gale/version"
	"github.com/aweris/gale/internal/config"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gale <command> [flags]",
		Short: "Gale is a tool to run Github Actions locally",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			client, err := dagger.Connect(cmd.Context(), dagger.WithLogOutput(cmd.OutOrStdout()))
			if err != nil {
				return err
			}

			config.SetClient(client)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close the client connection when the command is done.
			return config.Client().Close()
		},
	}

	return cmd
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
