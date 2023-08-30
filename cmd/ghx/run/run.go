package run

import (
	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/pkg/ghx"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	// TODO: improve DX here. It expects run-id as the first argument and config file should be stored in a specific location with a specific name
	return &cobra.Command{
		Use:   "run <run-id> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			var jr *core.JobRun

			dir := config.Client().Host().Directory(config.GhxRunDir(runID))

			jr, err := core.UnmarshalJobRunFromDir(cmd.Context(), dir)
			if err != nil {
				return err
			}

			runner, err := ghx.Plan(jr)
			if err != nil {
				return err
			}

			return runner.Run(cmd.Context())
		},
	}
}
