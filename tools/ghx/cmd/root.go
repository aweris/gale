package cmd

import (
	"fmt"
	"log"
	"os"

	"dagger.io/dagger"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aweris/gale/tools/ghx/cmd/run"
	"github.com/aweris/gale/tools/ghx/cmd/version"
	"github.com/aweris/gale/tools/ghx/config"
	"github.com/aweris/gale/tools/ghx/internal/fs"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var homeDir string

	cmd := &cobra.Command{
		Use:   "ghx",
		Short: "Github Actions Executor",
		Long:  "Github Actions Executor is a helper tool for gale to run workflows locally",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set home directory
			if err := fs.EnsureDir(homeDir); err != nil {
				return err
			}

			config.SetConfigHome(homeDir)

			// initialize dagger client and set it to config
			var opts []dagger.ClientOpt

			if os.Getenv("RUNNER_DEBUG") == "1" {
				opts = append(opts, dagger.WithLogOutput(os.Stdout))
			}

			client, err := dagger.Connect(cmd.Context(), opts...)
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

	cmd.PersistentFlags().StringVar(&homeDir, "home", "/home/runner/_temp/ghx", "home directory for ghx")

	bindEnv(cmd.Flags().Lookup("home"), "GHX_HOME")

	return cmd
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
