package run

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/model"
	"github.com/aweris/gale/tools/ghx/config"
	"github.com/aweris/gale/tools/ghx/internal/fs"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {

	// TODO: improve DX here. It expects run-id as the first argument and config file should be stored in a specific location with a specific name
	return &cobra.Command{
		Use:   "run <run-id> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			runID := args[0]

			var jc *model.JobConfig

			if err := fs.ReadJSONFile(filepath.Join(config.RunDir(runID), "config.json"), &jc); err != nil {
				return err
			}

			// TODO: handle execution of job config

			fmt.Printf("Running job %+v\n", jc)

			return nil
		},
	}
}
