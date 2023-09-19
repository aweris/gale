package list

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/gctx"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var (
		repo string
		opts gctx.LoadRepoOpts
		rc   *gctx.Context
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflows and jobs under it",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
			client, err := helpers.DefaultClient(cmd.Context())
			if err != nil {
				return err
			}

			config.SetClient(client)

			clientNoLog, err := helpers.NoLogClient(cmd.Context())
			if err != nil {
				return err
			}

			config.SetClientNoLog(clientNoLog)

			// Load context
			rc, err = gctx.Load(cmd.Context(), false)
			if err != nil {
				return err
			}

			// Load repository
			return rc.LoadRepo(repo, opts)
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close the client connection when the command is done.
			return config.Client().Close()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: add more information about the workflow like the trigger, etc.
			// TODO: maybe we could add better formatting for the output

			workflows, err := rc.LoadWorkflows()
			if err != nil {
				return err
			}

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
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&opts.WorkflowsDir, "workflows-dir", "", "directory to load workflows from. If empty, workflows will be loaded from the default directory.")

	return cmd
}
