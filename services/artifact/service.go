package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Service represents the artifact service
type Service interface {
	// CreateArtifactInNameContainer Create an entry for the artifact in the file container and return the container id
	CreateArtifactInNameContainer(runID string) (string, error)

	// UploadArtifactToFileContainer Upload file to the file container. Uploads are handled in chunks. The first chunk will
	// create the file and the subsequent chunks will append to the file.
	UploadArtifactToFileContainer(containerID, path, content string) error

	// PatchArtifactSize updates the size of the artifact to indicate we are done uploading. The uncompressed size
	// is used for display purposes however the implementation of the artifact service ignores it. It is only exist
	// to complete the artifact upload workflow.
	PatchArtifactSize(runID string)

	// ListArtifacts gets a list of all artifacts that are in a specific container and returns the container id and
	// the list of artifacts in the container
	ListArtifacts(runID string) (string, []string, error)

	// GetContainerItems gets container entries for the specific artifact
	GetContainerItems(containerID, path string) ([]string, error)

	// DownloadSingleArtifact downloads a single artifact from the specified container
	DownloadSingleArtifact(path string) (string, error)
}

var _ Service = new(LocalService)

type LocalService struct {
	path string // path to the artifact directory
}

func NewLocalService(path string) *LocalService {
	return &LocalService{path: path}
}

func (s *LocalService) CreateArtifactInNameContainer(runID string) (string, error) {
	path := filepath.Join(s.path, runID)

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	fmt.Printf("Created artifact directory %s\n", path)

	// we're using the runID as the containerID. Even it's no op return, it's a good practice to return the containerID
	// from service to keep conversion logic in the service layer. This way, the caller doesn't need to know how to
	// convert the runID to containerID

	return runID, nil
}

func (s *LocalService) UploadArtifactToFileContainer(containerID, path, content string) error {
	// Get the path to the artifact
	artifactPath := filepath.Join(s.path, containerID, filepath.Clean(path))

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0755); err != nil {
		return err
	}

	// Open the file in append mode
	file, err := os.OpenFile(artifactPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Append the data to the file
	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

func (s *LocalService) PatchArtifactSize(runID string) {
	fmt.Printf("Artifact upload complete for run %s\n", runID)
}

func (s *LocalService) ListArtifacts(runID string) (string, []string, error) {
	entries, err := fs.ReadDir(os.DirFS(s.path), runID)
	if err != nil {
		return "", nil, err
	}

	var artifacts []string

	for _, entry := range entries {
		artifacts = append(artifacts, entry.Name())
	}

	// we're using the runID as the containerID. Even it's no op return, it's a good practice to return the containerID
	// from service to keep conversion logic in the service layer. This way, the caller doesn't need to know how to
	// convert the runID to containerID

	return runID, artifacts, nil
}

func (s *LocalService) GetContainerItems(containerID, path string) ([]string, error) {
	artifactPath := filepath.Join(s.path, containerID, filepath.Clean(path))

	var files []string

	err := filepath.WalkDir(artifactPath, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		rel, err := filepath.Rel(filepath.Join(s.path, containerID), path)
		if err != nil {
			return err
		}

		// append the relative path to the files slice
		files = append(files, rel)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s *LocalService) DownloadSingleArtifact(path string) (string, error) {
	artifactPath := filepath.Join(s.path, path)

	content, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
