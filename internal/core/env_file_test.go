package core_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
)

// TODO: refactor this tests to run as a parameterized test. This is just a quick and dirty implementation.

func TestLocalEnvironmentFile_ReadData(t *testing.T) {
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

	dir, err := os.MkdirTemp("", "test_env_file")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "test_env_file.txt")

	// Test ensuring the file exists
	envFile, err := core.NewLocalEnvironmentFile(file)
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

	data, err := envFile.ReadData(context.Background())
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

func TestLocalEnvironmentFile_RawData(t *testing.T) {
	// Prepare a temporary test file.
	// Note: test data contains single key and key value pairs in a same file, however, this is happening in a real.
	testData := `
# This is a markdown

Some text here

## This is a header

More text here`

	dir, err := os.MkdirTemp("", "test_env_file")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)

	file := filepath.Join(dir, "test_env_file.txt")

	// Test ensuring the file exists
	envFile, err := core.NewLocalEnvironmentFile(file)
	if err != nil {
		t.Fatalf("Failed to create environment file: %v", err)
	}

	// Write test data to the file
	err = os.WriteFile(file, []byte(testData), 0600)
	if err != nil {
		t.Fatalf("Failed to write test data to file: %v", err)
	}

	// Test reading raw data from the file
	rawData, err := envFile.RawData(context.Background())
	if err != nil {
		t.Fatalf("Failed to read raw data from environment file: %v", err)
	}

	// Compare the expected data with the actual data
	if rawData != testData {
		t.Errorf("Expected raw data to be '%s', but got '%s'", testData, rawData)
	}
}

func TestDaggerEnvironmentFile_ReadData(t *testing.T) {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatalf("Failed to connect to dagger: %v", err)
	}
	defer client.Close()

	config.SetClient(client)

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

	dir := client.Directory().WithNewFile("test_env_file.txt", testData)

	// Test ensuring the file exists
	envFile := core.NewDaggerEnvironmentFile(dir.File("test_env_file.txt"))

	// Test reading key-value pairs and keys without values
	expectedData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
		"key4": "",
		"key5": "value4",
		"key6": "This is a multi-line\nvalue for key6.",
	}

	data, err := envFile.ReadData(context.Background())
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

func TestDaggerEnvironmentFile_RawData(t *testing.T) {
	ctx := context.Background()
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatalf("Failed to connect to dagger: %v", err)
	}
	defer client.Close()

	config.SetClient(client)

	// Prepare a temporary test file.
	// Note: test data contains single key and key value pairs in a same file, however, this is happening in a real.
	testData := `
# This is a markdown

Some text here

## This is a header

More text here`

	dir := client.Directory().WithNewFile("test_env_file.txt", testData)

	// Test ensuring the file exists
	envFile := core.NewDaggerEnvironmentFile(dir.File("test_env_file.txt"))

	// Test reading raw data from the file
	rawData, err := envFile.RawData(context.Background())
	if err != nil {
		t.Fatalf("Failed to read raw data from environment file: %v", err)
	}

	// Compare the expected data with the actual data
	if rawData != testData {
		t.Errorf("Expected raw data to be '%s', but got '%s'", testData, rawData)
	}
}
