package logger

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/gha"
	"github.com/aweris/gale/journal"
)

// TODO: This package is a work in progress. It will be replaced with a proper logging solution.
// For now, it is just a quick hack to get something working with the journal entries as well as workflow commands.

// Possible considerations:
// - Structured logging
// - Support for different log levels / verbosity / grouping.
// - Support for different sinks (stdout, stderr, file, etc.)
// - Support for different formats (text, json, etc.)

// Logger is the interface for logging.
type Logger interface {
	// Notice logs a message at the notice level.
	Notice(msg string)

	// Debug logs a message at the debug level.
	Debug(msg string)

	// Info logs a message at the info level.
	Info(msg string)

	// Warn logs a message at the warn level.
	Warn(msg string)

	// Error logs a message at the error level.
	Error(msg string)
}

var _ Logger = new(logger)

type logger struct {
	options options
}

// NewLogger creates a new logger instance with the given options.
func NewLogger(opts ...Opt) Logger {
	options := options{}
	for _, opt := range opts {
		opt(&options)
	}

	logger := &logger{options: options}

	handleJournalR(logger)

	return logger
}

func (l *logger) Notice(msg string) {
	fmt.Printf("%s\n", msg)
}

func (l *logger) Debug(msg string) {
	if l.options.verbose {
		fmt.Printf("%s\n", msg)
	}
}

func (l *logger) Info(msg string) {
	fmt.Printf("%s\n", msg)
}

func (l *logger) Warn(msg string) {
	fmt.Printf("%s\n", msg)
}

func (l *logger) Error(msg string) {
	fmt.Printf("%s\n", msg)
}

// handleJournalR handles logging for entries received from the journal reader.
func handleJournalR(logger *logger) {
	indentation := "\t"

	options := logger.options

	// no journal reader, nothing to do
	if options.journalR == nil {
		return
	}

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

			isCommand, command := gha.ParseCommand(entry.Message)
			if !isCommand {
				// if verbose is enabled, print entry itself, otherwise just print the message
				if options.verbose {
					logger.Info(indentation + entry.String())
				} else {
					logger.Info(indentation + entry.Message)
				}

				continue
			}

			// TODO: This is a quick hack to get something working. We should find a better way to handle this.

			switch command.Name {
			case "group":
				logger.Info(indentation + command.Value)
				indentation += "\t"
			case "endgroup":
				indentation = strings.TrimSuffix(indentation, "\t")
			case "notice":
				logger.Notice(indentation + command.Value)
			case "debug":
				logger.Debug(indentation + command.Value)
			case "warning":
				logger.Warn(indentation + command.Value)
			case "error":
				logger.Error(indentation + command.Value)
			default:
				// do nothing
			}
		}
	}()
}
