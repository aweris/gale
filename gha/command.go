package gha

import (
	"regexp"
)

var (
	commandRe  = regexp.MustCompile("^::([\\w-]+)(?:\\s+((?:[\\w-]+=[^,]+,)*[\\w-]+=[^,]+))??::(.*?)$")
	keyValueRe = regexp.MustCompile("([\\w-]+)=([^,]+)")
)

// TODO: add possible command names as constants

// Command represents a Workflow command to communicate with Runner.
// More info about the commands: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions
type Command struct {
	Name       string
	Parameters map[string]string
	Value      string
}

// ParseCommand parses a Workflow command string and returns a Command object. If the string is not a valid Workflow
// command, it returns false.
func ParseCommand(str string) (bool, *Command) {
	// Match the pattern against the string
	matches := commandRe.FindStringSubmatch(str)

	// if the string doesn't match the pattern, return false
	if len(matches) != 4 {
		return false, nil
	}

	// Extract the command keyword
	name := matches[1]

	// Extract the parameters (if any)
	parametersStr := matches[2]
	parameters := make(map[string]string)
	if parametersStr != "" {
		// Extract key-value pairs from parameters
		kvMatches := keyValueRe.FindAllStringSubmatch(parametersStr, -1)
		for _, match := range kvMatches {
			parameters[match[1]] = match[2]
		}
	}

	return true, &Command{
		Name:       name,
		Parameters: parameters,
		Value:      matches[3]}
}
