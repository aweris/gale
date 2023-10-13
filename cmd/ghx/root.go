package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/cmd/ghx/run"
	"github.com/aweris/gale/internal/cmd"
	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/fs"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var homeDir string

	command := &cobra.Command{
		Use:   "ghx",
		Short: "Github Actions Executor",
		Long:  "Github Actions Executor is a helper tool for gale to run workflows locally",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set home directory
			if err := fs.EnsureDir(homeDir); err != nil {
				return err
			}

			config.SetGhxHome(homeDir)

			return nil
		},
	}

	command.PersistentFlags().StringVar(&homeDir, "home", "/home/runner/_temp/ghx", "home directory for ghx")

	cmd.BindEnv(command.Flags().Lookup("home"), "GHX_HOME")

	return command
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(run.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}
