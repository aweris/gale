package main

import (
	"fmt"
	"os"

	"dagger.io/dagger"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/cmd/ghx/run"
	"github.com/aweris/gale/internal/cmd"
	"github.com/aweris/gale/internal/cmd/version"
	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/journal"
	"github.com/aweris/gale/internal/log"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	var homeDir string

	command := &cobra.Command{
		Use:   "ghx",
		Short: "Github Actions Executor",
		Long:  "Github Actions Executor is a helper tool for gale to run workflows locally",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// set home directory
			if err := fs.EnsureDir(homeDir); err != nil {
				return err
			}

			config.SetGhxHome(homeDir)

			// initialize dagger client and set it to config
			var opts []dagger.ClientOpt

			journalW, journalR := journal.Pipe()

			// Just print the same logger to stdout for now. We'll replace this with something interesting later.
			go logJournalEntries(journalR)

			opts = append(opts, dagger.WithLogOutput(journalW))

			client, err := dagger.Connect(cmd.Context(), opts...)
			if err != nil {
				return err
			}

			config.SetClient(client)

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			// Close the client connection when the command is done.
			return config.Client().Close()
		},
	}

	command.PersistentFlags().StringVar(&homeDir, "home", "/home/runner/_temp/ghx", "home directory for ghx")

	cmd.BindEnv(command.Flags().Lookup("home"), "GHX_HOME")

	return command
}

// Execute runs the command.
func Execute() {
	rootCmd := NewCommand()

	rootCmd.AddCommand(run.NewCommand())
	rootCmd.AddCommand(version.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v", err)
		os.Exit(1)
	}
}

func logJournalEntries(reader journal.Reader) {
	for {
		entry, ok := reader.ReadEntry()
		if !ok {
			break
		}

		// Skip internal entries if we're not in debug mode
		if entry.Type == journal.EntryTypeInternal && !config.Debug() {
			continue
		}

		isCommand, command := core.ParseCommand(entry.Message)
		if !isCommand {
			log.Info(entry.Message)
			continue
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
}
