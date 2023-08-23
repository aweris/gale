package preflight

import (
	"fmt"

	"github.com/aweris/gale/internal/log"
)

// Reporter represents anything that can report the validation results.
type Reporter interface {
	Report(t Task, result Result) error // Report reports the validation results with the check name and the result.
}

var _ Reporter = new(ConsoleReporter)

// ConsoleReporter is a reporter that reports the validation results to the console.
type ConsoleReporter struct{}

// NewConsoleReporter creates a new ConsoleReporter.
func NewConsoleReporter() *ConsoleReporter {
	return &ConsoleReporter{}
}

func (r *ConsoleReporter) Report(t Task, result Result) error {
	var status string

	switch result.Status {
	case Failed:
		status = colorize(string(result.Status), red)
	case Passed:
		status = colorize(string(result.Status), green)
	}

	log.Info(fmt.Sprintf("[%s] %s: %s", t.Type(), t.Name(), status))

	// if there is no message then return
	if len(result.Messages) == 0 {
		return nil
	}

	log.StartGroup()

	for _, msg := range result.Messages {
		switch msg.Level {
		case Error:
			log.Error(msg.Content)
		case Warning:
			log.Warn(msg.Content)
		case Info:
			log.Info(msg.Content)
		case Debug:
			log.Debug(msg.Content)
		}
	}

	log.EndGroup()

	return nil
}

// color is the color of the message.
type color int32

const (
	red   color = 31
	green color = 32
)

func colorize(s string, c color) string {
	return fmt.Sprintf("\033[0;%dm%s\033[0m", c, s)
}
