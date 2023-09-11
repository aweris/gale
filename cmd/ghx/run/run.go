package run

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/pkg/ghx"
)

var ErrWorkflowNotFound = errors.New("workflow not found")

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "run <workflow> <job> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load context
			rc, err := gctx.Load(cmd.Context())
			if err != nil {
				return err
			}

			// Load repository
			if err := rc.LoadCurrentRepo(); err != nil {
				return err
			}

			wf, ok := rc.Repo.Workflows[args[0]]
			if !ok {
				return ErrWorkflowNotFound
			}

			runner, err := ghx.Plan(wf, args[1])
			if err != nil {
				return err
			}

			return runner.Run(cmd.Context())
		},
	}

	return command
}
