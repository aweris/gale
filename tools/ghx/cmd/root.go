package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/tools/ghx/cmd/version"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "ghx",
		Short: "Github Actions Executor",
		Long:  "Github Actions Executor is a helper tool for gale to run workflows locally",
	}
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
