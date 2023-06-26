package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalService_CreateArtifactInNameContainer(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	service := NewLocalService(tmpDir)
	runID := "testRunID"

	containerID, err := service.CreateArtifactInNameContainer(runID)
	if err != nil {
		t.Errorf("Failed to create artifact container: %v", err)
	}

	// Verify the container directory exists
	containerPath := filepath.Join(tmpDir, runID)
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		t.Errorf("Container directory does not exist")
	}

	// Verify the container ID is correct
	if containerID != runID {
		t.Errorf("Expected container ID %s, but got %s", runID, containerID)
	}
}

func TestLocalService_UploadArtifactToFileContainer(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	service := NewLocalService(tmpDir)
	containerID := "testContainerID"
	filePath := "test/path/file.txt"
	content := "test content"

	err := service.UploadArtifactToFileContainer(containerID, filePath, content)
	if err != nil {
		t.Errorf("Failed to upload artifact: %v", err)
	}

	// Verify the artifact file exists
	artifactPath := filepath.Join(tmpDir, containerID, filePath)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		t.Errorf("Artifact file does not exist")
	}

	// Verify the content of the artifact file
	fileContent, err := os.ReadFile(artifactPath)
	if err != nil {
		t.Errorf("Failed to read artifact file: %v", err)
	}

	if string(fileContent) != content {
		t.Errorf("Expected artifact content %q, but got %q", content, string(fileContent))
	}
}

func TestLocalService_ListArtifacts(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	service := NewLocalService(tmpDir)
	runID := "testRunID"

	// Create some test artifacts
	artifacts := []string{"foo/bar.txt", "some/test.txt"}
	containerPath := filepath.Join(tmpDir, runID)
	for _, artifact := range artifacts {
		artifactPath := filepath.Join(containerPath, artifact)
		err := os.MkdirAll(filepath.Dir(artifactPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create artifact directory: %v", err)
		}
		err = os.WriteFile(artifactPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write artifact file: %v", err)
		}
	}

	// List artifacts
	listedRunID, listedArtifacts, err := service.ListArtifacts(runID)
	if err != nil {
		t.Errorf("Failed to list artifacts: %v", err)
	}

	// Verify the listed runID matches
	if listedRunID != runID {
		t.Errorf("Expected runID %s, but got %s", runID, listedRunID)
	}

	// Verify the listed artifacts match
	if len(listedArtifacts) != len(artifacts) {
		t.Errorf("Expected %d artifacts, but got %d", len(artifacts), len(listedArtifacts))
	}

	for _, artifact := range artifacts {
		found := false
		for _, listedArtifact := range listedArtifacts {
			// check artifact directory with artifact
			if filepath.Dir(artifact) == listedArtifact {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Artifact %s not found in the listed artifacts", artifact)
		}
	}
}

func TestLocalService_GetContainerItems(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	service := NewLocalService(tmpDir)
	containerID := "testContainerID"

	// Create some test artifacts
	artifacts := []string{"file1.txt", "file2.txt", "dir/file3.txt"}
	containerPath := filepath.Join(tmpDir, containerID)
	for _, artifact := range artifacts {
		artifactPath := filepath.Join(containerPath, artifact)
		err := os.MkdirAll(filepath.Dir(artifactPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create artifact directory: %v", err)
		}
		err = os.WriteFile(artifactPath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write artifact file: %v", err)
		}
	}

	// Get container items
	items, err := service.GetContainerItems(containerID, "")
	if err != nil {
		t.Errorf("Failed to get container items: %v", err)
	}

	// Verify the number of items
	expectedItems := len(artifacts)
	if len(items) != expectedItems {
		t.Errorf("Expected %d items, but got %d", expectedItems, len(items))
	}

	// Verify the items match
	for _, artifact := range artifacts {
		found := false
		for _, item := range items {
			if artifact == item {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Artifact %s not found in the container items", artifact)
		}
	}
}

func TestLocalService_DownloadSingleArtifact(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := os.TempDir()
	defer os.RemoveAll(tmpDir)

	service := NewLocalService(tmpDir)
	filePath := "test/path/file.txt"
	content := "test content"

	// Create the artifact file
	artifactPath := filepath.Join(tmpDir, filePath)
	err := os.MkdirAll(filepath.Dir(artifactPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create artifact directory: %v", err)
	}
	err = os.WriteFile(artifactPath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write artifact file: %v", err)
	}

	// Download the artifact
	downloadedContent, err := service.DownloadSingleArtifact(filePath)
	if err != nil {
		t.Errorf("Failed to download artifact: %v", err)
	}

	// Verify the downloaded content matches
	if downloadedContent != content {
		t.Errorf("Expected downloaded content %q, but got %q", content, downloadedContent)
	}
}
