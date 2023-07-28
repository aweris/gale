package core_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aweris/gale/internal/core"
)

func TestEnvironmentFile_ReadData(t *testing.T) {
	// Prepare a temporary test file.
	// Note: test data contains single key and key value pairs in a same file, however, this is happening in a real.
	testData := `
key1=value1
key2=value2
key3=value3
key4
key5=value4
key6<<END
This is a multi-line
value for key6.
END
`
	dir := os.TempDir()
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "test_env_file.txt")

	// Test ensuring the file exists
	envFile, err := core.NewEnvironmentFile(file)
	if err != nil {
		t.Fatalf("Failed to create environment file: %v", err)
	}

	// Write test data to the file
	err = os.WriteFile(file, []byte(testData), 0600)
	if err != nil {
		t.Fatalf("Failed to write test data to file: %v", err)
	}

	// Test reading key-value pairs and keys without values
	expectedData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "",
		"key5": "value4",
		"key6": "This is a multi-line\nvalue for key6.",
	}

	data, err := envFile.ReadData()
	if err != nil {
		t.Fatalf("Failed to read data from environment file: %v", err)
	}

	// Compare the expected data with the actual data
	if len(data) != len(expectedData) {
		t.Errorf("Expected %d key-value pairs, but got %d", len(expectedData), len(data))
	}

	for key, expectedValue := range expectedData {
		if value, ok := data[key]; !ok || value != expectedValue {
			t.Errorf("Expected value for key '%s' to be '%s', but got '%s'", key, expectedValue, value)
		}
	}
}

func TestEnvironmentFile_RawData(t *testing.T) {
	// Prepare a temporary test file.
	// Note: test data contains single key and key value pairs in a same file, however, this is happening in a real.
	testData := `
# This is a markdown

Some text here

## This is a header

More text here`

	dir := os.TempDir()
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "test_env_file.txt")

	// Test ensuring the file exists
	envFile, err := core.NewEnvironmentFile(file)
	if err != nil {
		t.Fatalf("Failed to create environment file: %v", err)
	}

	// Write test data to the file
	err = os.WriteFile(file, []byte(testData), 0600)
	if err != nil {
		t.Fatalf("Failed to write test data to file: %v", err)
	}

	// Test reading raw data from the file
	rawData, err := envFile.RawData()
	if err != nil {
		t.Fatalf("Failed to read raw data from environment file: %v", err)
	}

	// Compare the expected data with the actual data
	if string(rawData) != testData {
		t.Errorf("Expected raw data to be '%s', but got '%s'", testData, string(rawData))
	}
}
