package gctx

import (
	"dagger.io/dagger"
	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/log"

	"github.com/aweris/gale/internal/journal"
)

// LogFunc is the function that will be called when a new journal entry is received for default dagger client.
type LogFunc func(entry *journal.Entry)

type DaggerContext struct {
	Client *dagger.Client // Client is the dagger client to be used in the workflow.
}

func (c *Context) LoadDaggerContext() error {
	// initialize dagger client and set it to config
	var opts []dagger.ClientOpt

	journalW, journalR := journal.Pipe()

	// Just print the same logger to stdout for now. We'll replace this with something interesting later.
	go logJournal(journalR, logJournalEntries)

	opts = append(opts, dagger.WithLogOutput(journalW))

	client, err := dagger.Connect(c.Context, opts...)
	if err != nil {
		return err
	}

	c.Dagger = DaggerContext{Client: client}

	return nil
}

func logJournal(reader journal.Reader, logFn LogFunc) {
	for {
		entry, ok := reader.ReadEntry()
		if !ok {
			break
		}

		logFn(entry)
	}
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
