package logger

import (
	"fmt"

	"github.com/aweris/gale/journal"
)

// Logger is the interface for logging.
type Logger interface {
	// Info logs a message at the info level.
	Info(msg string)

	// TODO: add notice, debug, warning, error as well
}

var _ Logger = new(logger)

type logger struct {
	// TBD
}

func NewLogger(opts ...Opt) Logger {
	options := options{}
	for _, opt := range opts {
		opt(&options)
	}

	logger := &logger{}

	if options.journalR != nil {
		handleJournalR(options, logger)
	}

	return logger
}

// TODO: implement proper logging, for now just print to stdout

func (l *logger) Info(msg string) {
	fmt.Printf("[INFO] %s\n", msg)
}

// handleJournalR handles logging for entries received from the journal reader.
func handleJournalR(options options, logger *logger) {
	// Just print the same logger to stdout for now. We'll replace this with something interesting later.
	go func() {
		for {
			entry, ok := options.journalR.ReadEntry()
			if !ok {
				break
			}

			// skip internal dagger logs unless verbose is enabled
			if !options.verbose && entry.Type != journal.EntryTypeExecution {
				continue
			}

			// if verbose is enabled, print entry itself, otherwise just print the message
			if options.verbose {
				logger.Info(entry.String())
			} else {
				logger.Info(entry.Message)
			}
		}
	}()
}
