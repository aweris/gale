package run

import (
	"errors"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/idgen"
	"github.com/aweris/gale/pkg/ghx"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrJobNotFound      = errors.New("job not found")
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var workflowsDir string // workflowsDir is the directory to load workflows from.

	cmd := &cobra.Command{
		Use:   "run <workflow> <job> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// `gale` already mounts and configures the repository directory as working directory. So we can
			// look for repository information in the current directory.
			//
			// TODO: do not make this call here. Pass info from `gale` to `ghx` instead. This is just easier for now.
			repo, err := core.GetRepository("")
			if err != nil {
				return err
			}

			workflows, err := repo.LoadWorkflows(cmd.Context(), core.RepositoryLoadWorkflowOpts{WorkflowsDir: workflowsDir})
			if err != nil {
				return err
			}

			wf, ok := workflows[args[0]]
			if !ok {
				return ErrWorkflowNotFound
			}

			jm, ok := wf.Jobs[args[1]]
			if !ok {
				return ErrJobNotFound
			}

			_, err = idgen.GenerateWorkflowRunID()
			if err != nil {
				return err
			}

			jobRunID, err := idgen.GenerateJobRunID()
			if err != nil {
				return err
			}

			jr := core.NewJobRun(jobRunID, jm)
			if err != nil {
				return err
			}

			runner, err := ghx.Plan(&jr)
			if err != nil {
				return err
			}

			return runner.Run(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&workflowsDir, "workflows-dir", "", "directory to load workflows from. If empty, workflows will be loaded from the default directory.")

	bindEnv(cmd.Flags().Lookup("workflows-dir"), "GALE_WORKFLOWS_DIR")

	return cmd
}

func bindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
