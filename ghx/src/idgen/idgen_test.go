package idgen_test

import (
	"os"
	"testing"

	"ghx/context"
	"ghx/idgen"
)

func TestGenerateWorkflowRunID(t *testing.T) {
	tempdir, err := os.MkdirTemp("", "test-gen-workflow-run-id")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	// ensure that the test data directory is deleted after the test
	defer os.RemoveAll(tempdir)

	ctx := &context.Context{
		GhxConfig: context.GhxConfig{
			HomeDir: tempdir,
		},
	}

	// Test first workflow run ID generation
	wfRunID, err := idgen.GenerateWorkflowRunID(ctx)
	if err != nil {
		t.Fatalf("Error generating workflow run ID: %v", err)
	}

	if wfRunID != "1" {
		t.Errorf("Expected first workflow run ID to be 1, got %s", wfRunID)
	}

	// Test second workflow run ID generation
	wfRunID, err = idgen.GenerateWorkflowRunID(ctx)
	if err != nil {
		t.Fatalf("Error generating workflow run ID: %v", err)
	}

	if wfRunID != "2" {
		t.Errorf("Expected second workflow run ID to be 2, got %s", wfRunID)
	}
}

func TestGenerateJobRunID(t *testing.T) {
	tempdir, err := os.MkdirTemp("", "test-gen-job-run-id")
	if err != nil {
		t.Fatalf("Error creating temp dir: %v", err)
	}
	// ensure that the test data directory is deleted after the test
	defer os.RemoveAll(tempdir)

	ctx := &context.Context{
		GhxConfig: context.GhxConfig{
			HomeDir: tempdir,
		},
	}

	// Test first job run ID generation
	jobRunID, err := idgen.GenerateJobRunID(ctx)
	if err != nil {
		t.Fatalf("Error generating job run ID: %v", err)
	}

	if jobRunID != "1" {
		t.Errorf("Expected first job run ID to be 1, got %s", jobRunID)
	}

	// Test second job run ID generation
	jobRunID, err = idgen.GenerateJobRunID(ctx)
	if err != nil {
		t.Fatalf("Error generating job run ID: %v", err)
	}

	if jobRunID != "2" {
		t.Errorf("Expected second job run ID to be 2, got %s", jobRunID)
	}
}
