package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/cmd/list"
	"github.com/aweris/gale/cmd/run"
	"github.com/aweris/gale/cmd/version"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gale <command> [flags]",
		Short: "Gale is a tool to run Github Actions locally",
	}

	return cmd
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(list.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
