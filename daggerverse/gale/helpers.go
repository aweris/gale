package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

// root returns the root directory of the project.
func root() string {
	// get location of current file
	_, current, _, _ := runtime.Caller(0)

	println(current)

	return filepath.Join(filepath.Dir(current), "../..")
}

// root returns the root directory of the project.
func (h *Host) root(opts ...HostDirectoryOpts) *Directory {
	return h.Directory(root(), opts...)
}

// unmarshalContentsToYAML unmarshal the contents of the file as YAML into the given value.
func (f *File) unmarshalContentsToYAML(ctx context.Context, v interface{}) error {
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

// unmarshalContentsToJSON unmarshal the contents of the file as JSON into the given value.
func (f *File) unmarshalContentsToJSON(ctx context.Context, v interface{}) error {
	stdout, err := f.Contents(ctx)
	if err != nil {
		return fmt.Errorf("%w: failed to get file contents", err)
	}

	err = json.Unmarshal([]byte(stdout), v)
	if err != nil {
		return fmt.Errorf("%w: failed to unmarshal file contents", err)
	}

	return nil
}
