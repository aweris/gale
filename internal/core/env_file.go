package core

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/fs"
)

// Name of the environment files used by github actions with the names used in the documentation. These names are list
// of the possible environment files that can be used by github actions. They do not represent the actual file names.
const (
	EnvFileNameGithubEnv          = "GITHUB_ENV"
	EnvFileNameGithubPath         = "GITHUB_PATH"
	EnvFileNameGithubStepSummary  = "GITHUB_STEP_SUMMARY"
	EnvFileNameGithubActionOutput = "GITHUB_ACTION_OUTPUT"
)

// EnvironmentFile represents a generated temporary file that can be used to perform certain actions. This struct is
// containing the path of the file and logic to read the content of the file.
//
// See: https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#environment-files
type EnvironmentFile interface {
	// RawData returns the raw data of the environment file.
	RawData(ctx context.Context) (string, error)

	// ReadData reads the data of the environment file and returns a map of key value pairs. if the file doesn't contain
	// a key value pair, line will be considered as a key and value will be empty string.
	ReadData(ctx context.Context) (map[string]string, error)
}

type LocalEnvironmentFile struct {
	Path string
}

// NewLocalEnvironmentFile creates a new environment file from the given path.
func NewLocalEnvironmentFile(path string) (*LocalEnvironmentFile, error) {
	// ensure the file exists
	if err := fs.EnsureFile(path); err != nil {
		return nil, err
	}

	return &LocalEnvironmentFile{Path: path}, nil
}

func (f LocalEnvironmentFile) RawData(_ context.Context) (string, error) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (f LocalEnvironmentFile) ReadData(_ context.Context) (map[string]string, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return read(file)
}

type DaggerEnvironmentFile struct {
	File *dagger.File // File is the file of the environment file.
}

// NewDaggerEnvironmentFile creates a new environment file from the given dagger file.
func NewDaggerEnvironmentFile(file *dagger.File) *DaggerEnvironmentFile {
	return &DaggerEnvironmentFile{File: file}
}

func (f DaggerEnvironmentFile) RawData(ctx context.Context) (string, error) {
	return f.File.Contents(ctx)
}

func (f DaggerEnvironmentFile) ReadData(ctx context.Context) (map[string]string, error) {
	raw, err := f.RawData(ctx)
	if err != nil {
		return nil, err
	}

	return read(strings.NewReader(raw))
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
