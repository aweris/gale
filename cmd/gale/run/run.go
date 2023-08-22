package run

import (
	"dagger.io/dagger"
	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/pkg/gale"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var (
		runnerImage string       // runnerImage is the image used for running the actions.
		opts        gale.RunOpts // options for the run command
	)

	cmd := &cobra.Command{
		Use:          "run <workflow> <job> [flags]",
		Short:        "Run Github Actions by providing workflow and job name",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true, // don't print usage when error occurs
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			client, err := dagger.Connect(cmd.Context(), dagger.WithLogOutput(cmd.OutOrStdout()))
			if err != nil {
				return err
			}

			config.SetClient(client)

			if runnerImage != "" {
				config.SetRunnerImage(runnerImage)
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close the client connection when the command is done.
			return config.Client().Close()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := config.Client().Container().
				From(config.RunnerImage()).
				With(gale.ExecutionEnv(cmd.Context())).
				With(gale.Run(cmd.Context(), args[0], args[1], opts)).
				Sync(cmd.Context())

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&runnerImage, "runner", "", "runner image or path to Dockerfile to use for running the actions. If empty, the default runner image will be used.")
	cmd.Flags().StringVar(&opts.Repo, "repo", "", "owner/repo to load workflows from. If empty, repository information of the current directory will be used.")
	cmd.Flags().StringVar(&opts.Branch, "branch", "", "branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&opts.Tag, "tag", "", "tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&opts.WorkflowsDir, "workflows-dir", "", "directory to load workflows from. If empty, workflows will be loaded from the default directory.")
	cmd.Flags().StringToStringVar(&opts.Secrets, "secret", nil, "secrets to be used in the workflow. Format: --secret name=value")

	return cmd
}
