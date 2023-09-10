package helpers

import (
	"context"
	"os"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/journal"
)

// LogFunc is the function that will be called when a new journal entry is received for default dagger client.
type LogFunc func(entry *journal.Entry)

// NoopLogFunc is a log function that does nothing.
func NoopLogFunc(_ *journal.Entry) {}

// DefaultClient creates a new dagger client that logs to stdout.
func DefaultClient(ctx context.Context) (*dagger.Client, error) {
	return NewClient(ctx, nil)
}

// TODO: no log client is ok for now but what we really need is a client that logs to a file so we can use it for
//  debugging. However, we can add this on later iterations.

// NoLogClient creates a new dagger client that does not log anything.
func NoLogClient(ctx context.Context) (*dagger.Client, error) {
	return NewClient(ctx, NoopLogFunc)
}

// NewClient creates a new dagger client with given log function.
func NewClient(ctx context.Context, logFn LogFunc) (*dagger.Client, error) {
	// initialize dagger client and set it to config
	var opts []dagger.ClientOpt

	if logFn != nil {
		journalW, journalR := journal.Pipe()

		// Just print the same logger to stdout for now. We'll replace this with something interesting later.
		go logJournal(journalR, logFn)

		opts = append(opts, dagger.WithLogOutput(journalW))
	} else {
		opts = append(opts, dagger.WithLogOutput(os.Stdout))
	}

	client, err := dagger.Connect(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return client, nil
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
