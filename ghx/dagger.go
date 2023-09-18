package main

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/ghx/journal"
	"github.com/aweris/gale/internal/log"
)

func getDaggerClient(ctx context.Context) (*dagger.Client, error) {
	// initialize dagger client and set it to config
	var opts []dagger.ClientOpt

	journalW, journalR := journal.Pipe()

	// Just print the same logger to stdout for now. We'll replace this with something interesting later.
	go logJournal(journalR)

	opts = append(opts, dagger.WithLogOutput(journalW))

	return dagger.Connect(ctx, opts...)
}

func logJournal(reader journal.Reader) {
	cp := NewLoggingCommandProcessor()

	for {
		entry, ok := reader.ReadEntry()
		if !ok {
			break
		}

		if entry.Type == journal.EntryTypeInternal && os.Getenv("RUNNER_DEBUG") != "1" {
			return
		}

		// only process logging commands so it won't need context to process. It's okay to send nil context.
		if err := cp.ProcessOutput(nil, entry.Message); err != nil {
			log.Errorf("failed to process journal entry", "message", entry.Message, "error", err)
		}
	}
}
