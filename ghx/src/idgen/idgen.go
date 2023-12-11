package idgen

import (
	"path/filepath"
	"strconv"

	"github.com/aweris/gale/common/fs"

	"ghx/context"
)

const (
	metadataFile     = "idgen.json"
	keyWorkflowRunID = "workflow_run_id"
	keyJobRunID      = "job_run_id"
)

type counter map[string]int

// TODO: This is not concurrency safe. Need to use lock file or something similar to make it concurrency safe

// GenerateWorkflowRunID generates a unique workflow run id for the given repository
func GenerateWorkflowRunID(ctx *context.Context) (string, error) {
	path, err := ctx.GetMetadataPath()
	if err != nil {
		return "", err
	}

	dataPath := filepath.Join(path, metadataFile)

	return generateID(dataPath, keyWorkflowRunID)
}

// GenerateJobRunID generates a unique job run id for the given repository
func GenerateJobRunID(ctx *context.Context) (string, error) {
	path, err := ctx.GetMetadataPath()
	if err != nil {
		return "", err
	}

	dataPath := filepath.Join(path, metadataFile)

	return generateID(dataPath, keyJobRunID)
}

func generateID(dataPath, key string) (string, error) {
	err := fs.EnsureFile(dataPath)
	if err != nil {
		return "", err
	}

	var ids counter

	err = fs.ReadJSONFile(dataPath, &ids)
	if err != nil {
		return "", err
	}

	if ids == nil {
		ids = make(counter)
	}

	ids[key]++

	err = fs.WriteJSONFile(dataPath, ids)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(ids[key]), nil
}
