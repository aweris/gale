package main

import (
	"bufio"
	stdContext "context"
	"fmt"
	"io"
	"os"
	"strings"

	"ghx/context"
)

// Name of the environment files used by github actions with the names used in the documentation. These names are list
// of the possible environment files that can be used by github actions. They do not represent the actual file names.
const (
	EnvFileNameGithubEnv         = "GITHUB_ENV"
	EnvFileNameGithubPath        = "GITHUB_PATH"
	EnvFileNameGithubStepSummary = "GITHUB_STEP_SUMMARY"
	EnvFileNameGithubOutput      = "GITHUB_OUTPUT"
)

// EnvironmentFile represents a generated temporary file that can be used to perform certain actions. This struct is
// containing the path of the file and logic to read the content of the file.
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#environment-files
type EnvironmentFile interface {
	// Path returns the path of the environment file.
	Path() string

	// RawData returns the raw data of the environment file.
	RawData(ctx stdContext.Context) (string, error)

	// ReadData reads the data of the environment file and returns a map of key value pairs. if the file doesn't contain
	// a key value pair, line will be considered as a key and value will be empty string.
	ReadData(ctx stdContext.Context) (map[string]string, error)
}

// EnvironmentFiles represents a set of environment files that used by github actions. These files are temporary files
// that can be used to perform certain actions.
type EnvironmentFiles struct {
	Env         EnvironmentFile // Env is the environment file that holds the environment variables
	Path        EnvironmentFile // Path is the environment file that holds the path variables
	Outputs     EnvironmentFile // Outputs is the environment file that holds the outputs
	StepSummary EnvironmentFile // StepSummary is the environment file that holds the step summary
}

func (ef *EnvironmentFiles) Process(ctx *context.Context) error {
	env, err := ef.Env.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
		// FIXME: for now it's just reporting but we should use this as source of truth for step extra env
		if err := ctx.SetStepEnv(k, v); err != nil {
			return err
		}
	}

	paths, err := ef.Path.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	path := os.Getenv("PATH")

	for p := range paths {
		path = fmt.Sprintf("%s:%s", path, p)

		// FIXME: for now it's just reporting but we should use this as source of truth for step path
		if err := ctx.AddStepPath(p); err != nil {
			return err
		}
	}

	if err := os.Setenv("PATH", path); err != nil {
		return err
	}

	outputs, err := ef.Outputs.ReadData(ctx.Context)
	if err != nil {
		return err
	}

	for k, v := range outputs {
		ctx.SetStepOutput(k, v)
	}

	stepSummary, err := ef.StepSummary.RawData(ctx.Context)
	if err != nil {
		return err
	}

	ctx.SetStepSummary(stepSummary)

	return nil
}

func read(r io.Reader) (map[string]string, error) {
	keyValues := make(map[string]string)

	scanner := bufio.NewScanner(r)

	var (
		inMultiLineValue bool            // indicates if the scanner is currently processing a multi-line value
		currentKey       string          // current key of the multi-line value
		endMarker        string          // end marker of the multi-line value
		valueBuilder     strings.Builder // builder to build the multi-line value
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if the line contains "<<" indicating the start of a multi-line value
		if strings.Contains(line, "<<") {
			if inMultiLineValue {
				return nil, fmt.Errorf("unexpected '<<' in line: %s", line)
			}

			parts := strings.SplitN(line, "<<", 2)
			currentKey = strings.TrimSpace(parts[0])
			valueBuilder.Reset()
			inMultiLineValue = true

			// Extract the end marker from the line
			endMarker = strings.TrimSpace(parts[1])

			continue
		}

		// Check if there is active multi-line value processing
		if inMultiLineValue {
			// Check if the line is the end of the multi-line value
			if strings.TrimSpace(line) == endMarker {
				inMultiLineValue = false
				value := strings.TrimSpace(strings.TrimSuffix(valueBuilder.String(), "\n"))
				keyValues[currentKey] = value
			} else {
				valueBuilder.WriteString(line + "\n") // scanner removes the new line character, so we need to add it back
			}

			continue
		}

		// split key and value by "="
		parts := strings.SplitN(line, "=", 2)

		// if there is no "=" in the line, then it is a key without value (e.g. "path" values in GITHUB_PATH)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			keyValues[key] = value
		} else if len(parts) == 1 {
			key := strings.TrimSpace(parts[0])
			keyValues[key] = ""
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return keyValues, nil
}
