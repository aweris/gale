package journal

import (
	"strconv"
	"strings"
	"time"
)

// EntryType is the type of journal entry
type EntryType string

const (
	// EntryTypeInternal is a log entry from dagger itself , not from the container execution.
	EntryTypeInternal EntryType = "internal"

	// EntryTypeExecution is a log entry from the container execution.
	EntryTypeExecution EntryType = "execution"
)

// Entry is a single log entry from the journal
type Entry struct {
	Raw        string
	ID         int
	Index      int
	Type       EntryType
	ElapsedTme time.Duration
	Message    string
}

// String returns the raw log line
func (e *Entry) String() string {
	return e.Raw
}

// parseEntry parses a log line into an entry
func parseEntry(id int, line string) *Entry {
	// Create a new entry with the raw log line and the ID. We'll fill in the rest later.
	entry := &Entry{Raw: line, ID: id}

	var (
		index   int
		info    string
		message string
	)

	// To convert log lines into entries, we need to parse the log line into its parts. We'll do this by splitting
	// the line on spaces and then parsing each part. We'll then use the parts to fill in the entry.
	//
	// To make this easier, we'll make some assumptions about the log format. This will make the parsing easier according
	// to the following rules based on dagger log format:
	// - Message format is <index> <info> <message>
	// - The index always an integer suffixed with : (e.g. 1:, 2:, etc.)
	// - The info could be a duration or a string. If it's a duration, it will be surrounded by [] and will have
	// string representation of a time.Duration (e.g. [0.217s]), otherwise it will be alphanumeric
	// (e.g. exec, resolve, etc.)
	// - The message is the rest of the line, which could be empty
	// - The info and message are optional
	// - If the info is a duration, the message is execution log, otherwise it's an internal dagger log
	parts := strings.Split(line, " ")

	// TODO: assume we'll be dealing with valid log lines for now. Make this more robust later.

	index, _ = strconv.Atoi(strings.TrimSuffix(parts[0], ":"))

	if len(parts) > 1 {
		info = strings.TrimSpace(parts[1])

		if len(parts) > 2 {
			message = strings.Join(parts[2:], " ") // not using TrimSpace here because we want to preserve whitespace in the message
		}
	}

	// if the info is a duration, then this is an execution log
	if strings.HasPrefix(info, "[") && strings.HasSuffix(info, "s]") {
		duration, err := time.ParseDuration(strings.TrimSuffix(strings.TrimPrefix(info, "["), "]"))
		if err == nil {
			entry.Index = index
			entry.Type = EntryTypeExecution
			entry.ElapsedTme = duration
			entry.Message = message

			return entry
		}
	}

	// if we get here, then this is an internal log, so we'll fill in the entry accordingly.
	// we're not interested further breaking down the log information for now.
	entry.Index = index
	entry.Type = EntryTypeInternal
	entry.Message = strings.Join([]string{info, message}, " ")

	return entry
}
