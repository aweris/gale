package run

import (
	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/gale"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var (
		runnerImage string       // runnerImage is the image used for running the actions.
		debug       bool         // debug is the flag to enable debug mode.
		repo        string       // repo is the repository to load workflows from.
		branch      string       // branch is the branch to load workflows from.
		tag         string       // tag is the tag to load workflows from.
		opts        gale.RunOpts // options for the run command
	)

	cmd := &cobra.Command{
		Use:          "run <workflow> <job> [flags]",
		Short:        "Run Github Actions by providing workflow and job name",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true, // don't print usage when error occurs
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

			if runnerImage != "" {
				config.SetRunnerImage(runnerImage)
			}

			config.SetDebug(debug)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close the client connection when the command is done.
			return config.Client().Close()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := core.GetRepository(repo, core.GetRepositoryOpts{Branch: branch, Tag: tag})
			if err != nil {
				return err
			}

			// new gale instance
			gi := gale.New(repo)

			_, err = config.Client().Container().
				From(config.RunnerImage()).
				With(gi.ExecutionEnv(cmd.Context())).
				With(gi.Run(cmd.Context(), args[0], args[1], opts)).
				Sync(cmd.Context())

			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&runnerImage, "runner", "", "runner image or path to Dockerfile to use for running the actions. If empty, the default runner image will be used.")
	cmd.Flags().BoolVar(&debug, "debug", false, "enable debug mode")
	cmd.Flags().StringVar(&repo, "repo", "", "owner/repo to load workflows from. If empty, repository information of the current directory will be used.")
	cmd.Flags().StringVar(&branch, "branch", "", "branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&tag, "tag", "", "tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.")
	cmd.Flags().StringVar(&opts.WorkflowsDir, "workflows-dir", "", "directory to load workflows from. If empty, workflows will be loaded from the default directory.")
	cmd.Flags().StringToStringVar(&opts.Secrets, "secret", nil, "secrets to be used in the workflow. Format: --secret name=value")

	return cmd
}
