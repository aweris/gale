package run

import (
	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/gctx"
	"github.com/aweris/gale/internal/journal"
	"github.com/aweris/gale/internal/log"
	"github.com/aweris/gale/pkg/ghx"
	"github.com/spf13/cobra"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var job string

	command := &cobra.Command{
		Use:   "run <workflow> [flags]",
		Short: "Runs a job given run id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := helpers.NewClient(cmd.Context(), logJournalEntries)
			if err != nil {
				return err
			}

			// Load context
			ctx, err := gctx.Load(cmd.Context(), config.Debug(), client)
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
			result, _ := runner.Run(ctx)

			err = fs.WriteJSONFile("/home/runner/_temp/ghx/result.json", &result)
			if err != nil {
				return err
			}

			return nil
		},
	}

	command.Flags().StringVarP(&job, "job", "j", "", "job name to run")

	return command
}

func logJournalEntries(entry *journal.Entry) {
	// Skip internal entries if we're not in debug mode
	if entry.Type == journal.EntryTypeInternal && !config.Debug() {
		return
	}

	isCommand, command := core.ParseCommand(entry.Message)
	if !isCommand {
		log.Info(entry.Message)
		return
	}

	// TODO: We should extract these to common place, currently we're duplicating the code when we need to parse the commands

	// process only logging based commands and ignore the rest
	switch command.Name {
	case "group":
		log.Info(command.Value)
		log.StartGroup()
	case "endgroup":
		log.EndGroup()
	case "debug":
		log.Debug(command.Value)
	case "error":
		log.Errorf(command.Value, "file", command.Parameters["file"], "line", command.Parameters["line"], "col", command.Parameters["col"], "endLine", command.Parameters["endLine"], "endCol", command.Parameters["endCol"], "title", command.Parameters["title"])
	case "warning":
		log.Warnf(command.Value, "file", command.Parameters["file"], "line", command.Parameters["line"], "col", command.Parameters["col"], "endLine", command.Parameters["endLine"], "endCol", command.Parameters["endCol"], "title", command.Parameters["title"])
	case "notice":
		log.Noticef(command.Value, "file", command.Parameters["file"], "line", command.Parameters["line"], "col", command.Parameters["col"], "endLine", command.Parameters["endLine"], "endCol", command.Parameters["endCol"], "title", command.Parameters["title"])
	default:
		// do nothing
	}
}
