package list

import (
	"fmt"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/pkg/repository"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all workflows and jobs under the current repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := dagger.Connect(cmd.Context())
			if err != nil {
				return err
			}

			workflows, err := repository.LoadWorkflows(cmd.Context(), client)
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

	return cmd
}
