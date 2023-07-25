package list

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
		Use:   "list",
		Short: "List all workflows and jobs under the current repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := core.GetRepository(repo, getOpts)
			if err != nil {
				return err
			}

			workflows, err := repo.LoadWorkflows(cmd.Context(), loadOpts)
			if err != nil {
				return err
			}

			// TODO: add more information about the workflow like the trigger, etc.
			// TODO: maybe we could add better formatting for the output

			for _, workflow := range workflows {
				fmt.Printf("Workflow: ")
				if workflow.Name != workflow.Path {
					fmt.Printf("%s (path: %s)\n", workflow.Name, workflow.Path)
				} else {
					fmt.Printf("%s\n", workflow.Name)
				}

				fmt.Println("Jobs:")

				for job := range workflow.Jobs {
					fmt.Printf(" - %s\n", job)
				}

				fmt.Println("") // extra empty line
			}

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
