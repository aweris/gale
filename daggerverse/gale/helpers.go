package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

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

// workflowWalkFunc is the type of the function called for each workflow file visited by walkWorkflowDir.
type workflowWalkFunc func(ctx context.Context, path string, file *File) (bool, error)

// walkWorkflowDir walks the workflows directory and calls the given function for each workflow file. If walk function
// returns false, the walk stops.
func walkWorkflowDir(ctx context.Context, dir *Directory, path string, fn workflowWalkFunc) error {
	wd := dir.Directory(path)
	entries, err := wd.Entries(ctx)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			walk, err := fn(ctx, filepath.Join(path, entry), wd.File(entry))
			if err != nil {
				return err
			}

			if !walk {
				return nil
			}
		}
	}

	return nil
}

// mapToKV converts a map to a list of key=value strings. This is a temporary workaround pending a map support in
// Dagger.
func mapToKV(m map[string]string) []string {
	var kv []string
	for k, v := range m {
		kv = append(kv, k+"="+v)
	}
	return kv
}
