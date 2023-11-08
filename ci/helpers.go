package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// root returns the root directory of the project.
func root() string {
	// get location of current file
	_, current, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(current), "..")
}

// unmarshalContentsToYAML unmarshal the contents of the file as YAML into the given value.
func unmarshalContentsToYAML(ctx context.Context, f *File, v interface{}) error {
	stdout, err := f.Contents(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get file contents", err)
	}

	err = yaml.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal file contents", err)
	}

	return nil
}
