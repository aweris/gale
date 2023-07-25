package run

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/core"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {

	var (
		repo     string
		getOpts  core.GetRepositoryOpts
		loadOpts core.RepositoryLoadWorkflowOpts
	)

	cmd := &cobra.Command{
		Use:          "run <workflow> <job> [flags]",
		Short:        "Run Github Actions by providing workflow and job name",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true, // don't print usage when error occurs
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := core.GetRepository(repo, getOpts)
			if err != nil {
				return err
			}

			workflows, err := repo.LoadWorkflows(cmd.Context(), loadOpts)
			if err != nil {
				return err
			}

			workflow, ok := workflows[args[0]]
			if !ok {
				return fmt.Errorf("workflow %s not found", args[0])
			}

			_, ok = workflow.Jobs[args[1]]
			if !ok {
				return fmt.Errorf("job %s not found in workflow %s", args[1], args[0])
			}

			fmt.Printf("Running workflow %s, job %s\n", args[0], args[1])

			return nil
		},
	}

	cmd.Flags().StringVar(&repo, "repo", "", "owner/repo to load workflows from. If empty, repository information of the current directory will be used.")
	cmd.Flags().StringVar(&getOpts.Branch, "branch", "", "branch to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&getOpts.Tag, "tag", "", "tag to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&getOpts.Commit, "commit", "", "commit to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.")
	cmd.Flags().StringVar(&loadOpts.Path, "path", "", "path of the workflows directory. If empty, default path .github/workflows will be used.")

	return cmd
}
