package main

import (
	"context"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
)

// NewDaggerEnvironmentFiles creates a new environment files in the given directory path.
func NewDaggerEnvironmentFiles(dir string, client *dagger.Client) (*dagger.Directory, *EnvironmentFiles) {
	// create a new directory with all empty files in it
	emptyDir := client.Directory().
		WithNewFile("env", "").
		WithNewFile("path", "").
		WithNewFile("outputs", "").
		WithNewFile("step_summary", "")

	// create environment files using the empty files
	files := &EnvironmentFiles{
		Env:         NewDaggerEnvironmentFile(filepath.Join(dir, "env"), emptyDir.File("env")),
		Path:        NewDaggerEnvironmentFile(filepath.Join(dir, "path"), emptyDir.File("path")),
		Outputs:     NewDaggerEnvironmentFile(filepath.Join(dir, "outputs"), emptyDir.File("outputs")),
		StepSummary: NewDaggerEnvironmentFile(filepath.Join(dir, "step_summary"), emptyDir.File("step_summary")),
	}

	return emptyDir, files
}

// DaggerEnvironmentFile represents an environment file that is located in the dagger file.
type DaggerEnvironmentFile struct {
	path string
	File *dagger.File // File is the file of the environment file.
}

// NewDaggerEnvironmentFile creates a new environment file from the given dagger file.
func NewDaggerEnvironmentFile(path string, file *dagger.File) *DaggerEnvironmentFile {
	return &DaggerEnvironmentFile{path: path, File: file}
}

func (f DaggerEnvironmentFile) Path() string {
	return f.path
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
