package log

import (
	"fmt"
	"os"
	"strings"
)

const (
	groupStart = "┏ "
	groupMid   = "┃ "
	groupEnd   = "┗ "

	LevelDebug  = "debug"
	LevelWarn   = "warn"
	LevelErr    = "error"
	LevelNotice = "notice"
)

type Logger struct {
	groups []string
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) StartGroup() {
	l.log(groupStart, "", "")
	l.groups = append(l.groups, groupMid)
}

func (l *Logger) EndGroup() {
	if len(l.groups) > 0 {
		l.groups = l.groups[:len(l.groups)-1]
	}

	l.log(groupEnd, "", "")
}

func (l *Logger) Info(message string) {
	l.log("", "", message)
}

func (l *Logger) Infof(message string, keyvals ...interface{}) {
	l.logf("", message, keyvals...)
}

func (l *Logger) Debug(message string) {
	if os.Getenv("RUNNER_DEBUG") == "1" {
		l.log("", LevelDebug, message)
	}
}

func (l *Logger) Debugf(message string, keyvals ...interface{}) {
	if os.Getenv("RUNNER_DEBUG") == "1" {
		l.logf(LevelDebug, message, keyvals...)
	}
}

func (l *Logger) Warn(message string) {
	l.log("", LevelWarn, message)
}

func (l *Logger) Warnf(message string, keyvals ...interface{}) {
	l.logf(LevelWarn, message, keyvals...)
}

func (l *Logger) Error(message string) {
	l.log("", LevelErr, message)
}

func (l *Logger) Errorf(message string, keyvals ...interface{}) {
	l.logf(LevelErr, message, keyvals...)
}

func (l *Logger) Notice(message string) {
	l.log("", LevelNotice, message)
}

func (l *Logger) Noticef(message string, keyvals ...interface{}) {
	l.logf(LevelNotice, message, keyvals...)
}

func (l *Logger) logf(level, message string, keyvals ...interface{}) {
	var args []string

	for i := 0; i < len(keyvals); i += 2 {
		if keyvals[i+1] != "" && keyvals[i+1] != nil {
			args = append(args, fmt.Sprintf("%s=%v", keyvals[i], keyvals[i+1]))
		}
	}

	l.log("", level, fmt.Sprintf("%s %s", message, strings.Join(args, ",")))
}

func (l *Logger) log(prefix, level, message string) {
	sb := strings.Builder{}

	if len(l.groups) > 0 {
		sb.WriteString(strings.Join(l.groups, ""))
	}

	if prefix != "" {
		sb.WriteString(prefix)
	}

	if level != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", level))
	}

	sb.WriteString(message)

	fmt.Println(sb.String())
}
