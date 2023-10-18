package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/context"
)

var (
	commandReColon = regexp.MustCompile(`^::([\w-]+)(?:\s+((?:[\w-]+=[^,]+,)*[\w-]+=[^,]+))??::(.*?)$`)
	commandReHash  = regexp.MustCompile(`##\[(\S+)([^]]*)](.*)?$`) //
)

// WorkflowCommand represents a Workflow command to communicate with Runner.
// More info about the commands: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type WorkflowCommand struct {
	Name       string
	Parameters map[string]string
	Value      string
}

type CommandName string

const (
	CommandNameGroup      CommandName = "group"
	CommandNameEndGroup   CommandName = "endgroup"
	CommandNameDebug      CommandName = "debug"
	CommandNameError      CommandName = "error"
	CommandNameWarning    CommandName = "warning"
	CommandNameNotice     CommandName = "notice"
	CommandNameSetEnv     CommandName = "set-env"
	CommandNameSetOutput  CommandName = "set-output"
	CommandNameSaveState  CommandName = "save-state"
	CommandNameAddMask    CommandName = "add-mask"
	CommandNameAddMatcher CommandName = "add-matcher"
	CommandNameAddPath    CommandName = "add-path"
)

type CommandProcessor struct {
	exclude map[CommandName]bool // exclude is a map of command names to exclude from processing
}

// NewLoggingCommandProcessor creates a new command processor that processes logging commands.
func NewLoggingCommandProcessor() *CommandProcessor {
	return NewCommandProcessor(
		CommandNameSetEnv,
		CommandNameSetOutput,
		CommandNameSaveState,
		CommandNameAddMask,
		CommandNameAddMatcher,
		CommandNameAddPath,
	)
}

// NewEnvCommandsProcessor returns a new command processor that processes commands that manipulate execution environment.
func NewEnvCommandsProcessor() *CommandProcessor {
	return NewCommandProcessor(
		CommandNameGroup,
		CommandNameEndGroup,
		CommandNameDebug,
		CommandNameError,
		CommandNameWarning,
		CommandNameNotice,
	)
}

func NewCommandProcessor(excluded ...CommandName) *CommandProcessor {
	exclude := make(map[CommandName]bool, len(excluded))

	for _, name := range excluded {
		exclude[name] = true
	}

	return &CommandProcessor{exclude: exclude}
}

func (p *CommandProcessor) ProcessOutput(ctx *context.Context, output string) error {
	isCmd, cmd := parseCommand(output)
	if !isCmd {
		log.Info(output)

		return nil
	}

	if p.exclude[CommandName(cmd.Name)] {
		return nil
	}

	switch CommandName(cmd.Name) {
	case CommandNameGroup:
		log.Info(cmd.Value)
		log.StartGroup()
	case CommandNameEndGroup:
		log.EndGroup()
	case CommandNameDebug:
		log.Debug(cmd.Value)
	case CommandNameError:
		log.Errorf(cmd.Value, "file", cmd.Parameters["file"], "line", cmd.Parameters["line"], "col", cmd.Parameters["col"], "endLine", cmd.Parameters["endLine"], "endCol", cmd.Parameters["endCol"], "title", cmd.Parameters["title"])
	case CommandNameWarning:
		log.Warnf(cmd.Value, "file", cmd.Parameters["file"], "line", cmd.Parameters["line"], "col", cmd.Parameters["col"], "endLine", cmd.Parameters["endLine"], "endCol", cmd.Parameters["endCol"], "title", cmd.Parameters["title"])
	case CommandNameNotice:
		log.Noticef(cmd.Value, "file", cmd.Parameters["file"], "line", cmd.Parameters["line"], "col", cmd.Parameters["col"], "endLine", cmd.Parameters["endLine"], "endCol", cmd.Parameters["endCol"], "title", cmd.Parameters["title"])
	case CommandNameSetEnv:
		if err := os.Setenv(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
		// FIXME: for now it's just reporting but we should use this as source of truth for step extra env
		if err := ctx.SetStepEnv(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
	case CommandNameSetOutput:
		if err := ctx.SetStepOutput(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
	case CommandNameSaveState:
		if err := ctx.SetStepState(cmd.Parameters["name"], cmd.Value); err != nil {
			return err
		}
	case CommandNameAddMask:
		log.Info(cmd.Value)
	case CommandNameAddMatcher:
		log.Info(cmd.Value)
	case CommandNameAddPath:
		// FIXME: for now it's just reporting but we should use this as source of truth for step path
		if err := ctx.AddStepPath(cmd.Value); err != nil {
			return err
		}
		path := os.Getenv("PATH")
		path = fmt.Sprintf("%s:%s", path, cmd.Value)
		if err := os.Setenv("PATH", path); err != nil {
			return err
		}
	}

	return nil
}

// parseCommand parses a Workflow command string and returns a Command object. If the string is not a valid Workflow
// command, it returns false.
func parseCommand(str string) (bool, *WorkflowCommand) {
	var command WorkflowCommand

	if matches := commandReColon.FindStringSubmatch(str); matches != nil {
		// Extract the command keyword
		command.Name = matches[1]
		command.Parameters = parseParameters(matches[2], ",")
		command.Value = matches[3]
	} else if matches := commandReHash.FindStringSubmatch(str); matches != nil {
		// Extract the command keyword
		command.Name = matches[1]
		command.Parameters = parseParameters(matches[2], ";")
		command.Value = matches[3]
	} else {
		return false, nil
	}

	return true, &command
}

func parseParameters(parametersStr string, separator string) map[string]string {
	parameters := make(map[string]string)

	if parametersStr == "" {
		return parameters
	}

	for _, parameter := range strings.Split(parametersStr, separator) {
		parts := strings.SplitN(parameter, "=", 2)
		if len(parts) == 2 {
			parameters[parts[0]] = parts[1]
		}
	}

	return parameters
}
