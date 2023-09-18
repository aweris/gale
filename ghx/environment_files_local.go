package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aweris/gale/internal/fs"
)

// NewLocalEnvironmentFiles creates a new environment files in the given directory path.
func NewLocalEnvironmentFiles(dir string) (*EnvironmentFiles, error) {
	files := &EnvironmentFiles{}

	if err := fs.EnsureDir(dir); err != nil {
		return nil, err
	}

	env, err := NewLocalEnvironmentFile(filepath.Join(dir, "env"))
	if err != nil {
		return nil, err
	}

	files.Env = env

	path, err := NewLocalEnvironmentFile(filepath.Join(dir, "path"))
	if err != nil {
		return nil, err
	}

	files.Path = path

	outputs, err := NewLocalEnvironmentFile(filepath.Join(dir, "outputs"))
	if err != nil {
		return nil, err
	}

	files.Outputs = outputs

	summary, err := NewLocalEnvironmentFile(filepath.Join(dir, "step_summary"))
	if err != nil {
		return nil, err
	}

	files.StepSummary = summary

	return files, nil
}

// LocalEnvironmentFile represents a local environment file that is located in the filesystem.
type LocalEnvironmentFile struct {
	path string
}

// NewLocalEnvironmentFile creates a new environment file from the given path.
func NewLocalEnvironmentFile(path string) (*LocalEnvironmentFile, error) {
	// ensure the file exists
	if err := fs.EnsureFile(path); err != nil {
		return nil, err
	}

	return &LocalEnvironmentFile{path: path}, nil
}

// Path returns the path of the environment file.
func (f LocalEnvironmentFile) Path() string {
	return f.path
}

func (f LocalEnvironmentFile) RawData(_ context.Context) (string, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (f LocalEnvironmentFile) ReadData(_ context.Context) (map[string]string, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return read(file)
}
