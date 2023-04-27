package journal

import (
	"strconv"
	"strings"
	"time"
)

// EntryType is the type of a journal entry
type EntryType string

const (
	// EntryTypeInternal is a log entry from dagger itself , not from the container execution.
	EntryTypeInternal EntryType = "internal"

	// EntryTypeExecution is a log entry from the container execution.
	EntryTypeExecution EntryType = "execution"

	// EntryTypeDone is a log entry notifying that the container execution is done for the given index.
	EntryTypeDone EntryType = "done"
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

func (e *Entry) String() string {
	return e.Raw
}

func parseEntry(id int, line string) *Entry {
	// Create a new entry with the raw log line and the ID. We'll fill in the rest later.
	entry := &Entry{Raw: line, ID: id}

	var (
		index   int
		info    string
		message string
	)

	// Log lines are space-delimited, so split the line into parts
	//
	// Our assumptions are:
	// - Message format is <index> <info> <message>
	// - The index is always an integer prefixed with a #
	// - The info is could be a float or a string. If it's a float, it's the elapsed time. Otherwise, it's part of the message or standalone info.
	// - The message is the rest of the line
	// - The info and message are optional
	parts := strings.Split(line, " ")

	// TODO: assume we'll be dealing with valid log lines for now. Make this more robust later.
	index, _ = strconv.Atoi(strings.TrimPrefix(parts[0], "#"))

	if len(parts) > 1 {
		info = strings.TrimSpace(parts[1])

		if len(parts) > 2 {
			message = strings.TrimSpace(strings.Join(parts[2:], " "))
		}
	}

	// If the info is a float, it's an execution log with an elapsed time
	// Example:	#3 0.217 ::debug::ref = 'undefined'
	if elapsed, err := strconv.ParseFloat(info, 64); err == nil {
		// The info is a float, which means it's an execution log
		entry.Index = index
		entry.Type = EntryTypeExecution
		entry.ElapsedTme = time.Duration(int64(elapsed * float64(time.Second)))
		entry.Message = message

		return entry
	}

	// If the info is "DONE", it's a log notifying that index group is done
	// Example: #1 DONE 0.0s
	if info == "DONE" {
		// The index group is done, and the next info is the total elapsed time as duration string
		duration, _ := time.ParseDuration(strings.TrimSpace(message))

		entry.Index = index
		entry.Type = EntryTypeDone
		entry.ElapsedTme = duration

		return entry
	}

	// The info is alphanumeric, which means it's an internal log
	// Example:
	// 	#1
	// 	#3 CACHED
	// 	#5 host.directory /home/aweris/.local/share/gale/ubuntu-latest
	entry.Index = index
	entry.Type = EntryTypeInternal
	entry.Message = strings.Join([]string{info, message}, " ")

	return entry
}
