package utils_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/aweris/gale/pkg/utils"
)

// TestMultipartFileWriter tests the functionality of MultipartFileWriter.
func TestMultipartFileWriter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := createTempDir(t)
	defer os.RemoveAll(tempDir)

	// Initialize the MultipartFileWriter with the temporary directory
	writer, err := utils.NewMultipartFileWriter(tempDir)
	if err != nil {
		t.Fatalf("Failed to create MultipartFileWriter: %v", err)
	}

	// Test Write and Merge operations
	t.Run("Write and Merge", func(t *testing.T) {
		// Test data
		filePath := "test.txt"
		data := []byte("Hello, World!")

		// Write the data in chunks
		err := writer.Write(filePath, 0, bytes.NewReader(data[:5]))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		err = writer.Write(filePath, 5, bytes.NewReader(data[5:]))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		// Merge the parts
		err = writer.Merge()
		if err != nil {
			t.Fatalf("Failed to merge parts: %v", err)
		}

		// Verify the merged file
		mergedFilePath := filepath.Join(tempDir, filePath)
		verifyFileContent(t, mergedFilePath, data)
	})

	// Test merging multiple files
	t.Run("Merge multiple files", func(t *testing.T) {
		// Test data
		filePaths := []string{"file1.txt", "file2.txt"}
		data1 := []byte("File 1")
		data2 := []byte("File 2")

		// Write the data for file1
		err := writer.Write(filePaths[0], 0, bytes.NewReader(data1))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		// Write the data for file2
		err = writer.Write(filePaths[1], 0, bytes.NewReader(data2))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		// Merge the parts
		err = writer.Merge()
		if err != nil {
			t.Fatalf("Failed to merge parts: %v", err)
		}

		// Verify the merged files
		for _, filePath := range filePaths {
			mergedFilePath := filepath.Join(tempDir, filePath)
			switch filePath {
			case filePaths[0]:
				verifyFileContent(t, mergedFilePath, data1)
			case filePaths[1]:
				verifyFileContent(t, mergedFilePath, data2)
			}
		}
	})

	// Test merging files with different offsets
	t.Run("Merge files with different offsets", func(t *testing.T) {
		// Test data
		filePath := "test.txt"
		data := []byte("Hello, World!")

		// Write the data in different offsets
		err := writer.Write(filePath, 5, bytes.NewReader(data[5:]))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		err = writer.Write(filePath, 0, bytes.NewReader(data[:5]))
		if err != nil {
			t.Fatalf("Failed to write part: %v", err)
		}

		// Merge the parts
		err = writer.Merge()
		if err != nil {
			t.Fatalf("Failed to merge parts: %v", err)
		}

		// Verify the merged file
		mergedFilePath := filepath.Join(tempDir, filePath)
		verifyFileContent(t, mergedFilePath, data)
	})
}

// createTempDir creates a temporary directory for testing and returns its path.
func createTempDir(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "multipart_file_writer_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	return tempDir
}

// verifyFileContent reads the content of the file at the specified path and verifies it against the expected data.
func verifyFileContent(t *testing.T, filePath string, expectedData []byte) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", filePath, err)
	}

	if !bytes.Equal(data, expectedData) {
		t.Errorf("File content mismatch. Expected: %s, Actual: %s", expectedData, data)
	}
}
