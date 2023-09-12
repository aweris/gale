package run

import (
	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/pkg/ghx"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "run <workflow> <job> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load context
			rc, err := gctx.Load(cmd.Context(), config.Debug())
			if err != nil {
				return err
			}

			// Load repository
			if err := rc.LoadCurrentRepo(); err != nil {
				return err
			}

			runner, err := ghx.Plan(rc, args[0], args[1])
			if err != nil {
				return err
			}

			return runner.Run()
		},
	}

	return command
}
