package context

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/aweris/gale/common/fs"
)

// GetMetadataPath returns the path of the metadata path. If the path does not exist, it creates it. This path assumes
// zenith module set a cache for metadata, otherwise it will be empty every time it runs.
func (c *Context) GetMetadataPath() (string, error) {
	return EnsureDir(c.GhxConfig.MetadataDir)
}

// GetActionsPath returns the path of the custom action repositories clones and stores. If the path does not exist, it
// creates it.
func (c *Context) GetActionsPath() (string, error) {
	return EnsureDir(c.GhxConfig.ActionsDir)
}

// GetSecretsPath returns the path of the secrets.json file containing the secrets. If the path does not exist, it
// creates it.
func (c *Context) GetSecretsPath() (string, error) {
	file := filepath.Join(c.GhxConfig.HomeDir, "secrets", "secrets.json")

	// ensure file and directory exists
	if err := fs.EnsureFile(file); err != nil {
		return "", err
	}

	// if file content is empty, write empty json object to avoid json unmarshal error
	stat, err := os.Stat(file)
	if err != nil {
		return "", err
	}

	if stat.Size() == 0 {
		fs.WriteJSONFile(file, map[string]string{})
	}

	return file, nil
}

// GetWorkflowRunPath returns the path of the current workflow run path. If the path does not exist, it creates it. If
// the workflow run is not set, it returns an error.
func (c *Context) GetWorkflowRunPath() (string, error) {
	return EnsureDir(c.GhxConfig.HomeDir, "run")
}

// GetJobRunPath returns the path of the current job run path. If the path does not exist, it creates it. If the job run
// is not set, it returns an error.
func (c *Context) GetJobRunPath() (string, error) {
	if c.Execution.JobRun == nil {
		return "", errors.New("no job is set")
	}

	return EnsureDir(c.GhxConfig.HomeDir, "run", "jobs", c.Execution.JobRun.Job.ID)
}

// GetStepRunPath returns the path of the current step run path. If the path does not exist, it creates it. If the step
// run is not set, it returns an error.
func (c *Context) GetStepRunPath() (string, error) {
	if c.Execution.StepRun == nil {
		return "", errors.New("no step is set")
	}

	return EnsureDir(c.GhxConfig.HomeDir, "run", "jobs", c.Execution.JobRun.Job.ID, "steps", c.Execution.StepRun.Step.ID)
}

// EnsureDir return the joined path and ensures that the directory exists. and returns the joined path.
func EnsureDir(path ...string) (string, error) {
	joined := filepath.Join(path...)

	if err := fs.EnsureDir(joined); err != nil {
		return "", err
	}

	return joined, nil
}
