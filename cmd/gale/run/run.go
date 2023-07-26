package run

import (
	"github.com/aweris/gale/internal/config"
	"github.com/spf13/cobra"

	"github.com/aweris/gale/pkg/gale"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var opts gale.RunOpts // options for the run command

	cmd := &cobra.Command{
		Use:          "run <workflow> <job> [flags]",
		Short:        "Run Github Actions by providing workflow and job name",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true, // don't print usage when error occurs
		RunE: func(cmd *cobra.Command, args []string) error {
			client := config.Client()

			client.Container().From("alpine:latest").With(gale.Run(cmd.Context(), args[0], args[1], opts))

			return nil
		},
	}

	cmd.Flags().StringVar(&opts.Repo, "repo", "", "owner/repo to load workflows from. If empty, repository information of the current directory will be used.")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "branch to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "tag to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&opts.Commit, "commit", "", "commit to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&opts.WorkflowsDir, "workflows-dir", "", "directory to load workflows from. If empty, workflows will be loaded from the default directory.")

	return cmd
}
