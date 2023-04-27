package logger

import "github.com/aweris/gale/journal"

// options is the set of options for the logger.
type options struct {
	// verbose is true if verbose logging is enabled including debug and internal dagger logs.
	verbose bool

	// JournalR is the journal reader to use for logging.
	journalR journal.Reader
}

// Opt is a function that configures the logger.
type Opt func(*options)

// WithVerbose returns an option that enables verbose logging.
func WithVerbose(verbose bool) Opt {
	return func(o *options) {
		o.verbose = verbose
	}
}

// WithJournalR returns an option that sets the journal reader to logger journal entries as they are received.
func WithJournalR(journalR journal.Reader) Opt {
	return func(o *options) {
		o.journalR = journalR
	}
}
