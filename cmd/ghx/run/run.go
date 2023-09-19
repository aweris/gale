package run

import (
	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/pkg/ghx"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var job string

	command := &cobra.Command{
		Use:   "run <workflow> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load context
			ctx, err := gctx.Load(cmd.Context(), config.Debug())
			if err != nil {
				return err
			}

			// Load repository
			if err := ctx.LoadCurrentRepo(); err != nil {
				return err
			}

			// Load workflow
			workflows, err := ctx.LoadWorkflows()
			if err != nil {
				return err
			}

			wf, ok := workflows[args[0]]
			if !ok {
				return ghx.ErrWorkflowNotFound
			}

			// Create task runner for the workflow
			runner, err := ghx.Plan(wf, job)
			if err != nil {
				return err
			}

			// Run the workflow
			_, _, err = runner.Run(ctx)
			if err != nil {
				return err
			}

			return nil
		},
	}

	command.Flags().StringVarP(&job, "job", "j", "", "job name to run")

	return command
}
