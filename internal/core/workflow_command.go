package core

import (
	"regexp"
	"strings"
)

var (
	commandReColon = regexp.MustCompile(`^::([\w-]+)(?:\s+((?:[\w-]+=[^,]+,)*[\w-]+=[^,]+))??::(.*?)$`)
	commandReHash  = regexp.MustCompile(`##\[(\S+)([^]]*)](.*)?$`) //
)

// TODO: add possible command names as constants

// WorkflowCommand represents a Workflow command to communicate with Runner.
// More info about the commands: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type WorkflowCommand struct {
	Name       string
	Parameters map[string]string
	Value      string
}

// ParseCommand parses a Workflow command string and returns a Command object. If the string is not a valid Workflow
// command, it returns false.
func ParseCommand(str string) (bool, *WorkflowCommand) {
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
