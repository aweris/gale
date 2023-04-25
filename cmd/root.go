package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/cmd/build"
	"github.com/aweris/gale/cmd/run"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use: "gale",
	}
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(build.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
