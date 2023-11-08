package main

import (
	"context"
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

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

// unmarshalContentsToJSON unmarshal the contents of the file as JSON into the given value.
func unmarshalContentsToJSON(ctx context.Context, f *File, v interface{}) error {
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
