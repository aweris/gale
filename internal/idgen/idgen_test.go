package idgen_test

// FIXME: This test is broken after the refactoring. Need to fix it.

/*
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/idgen"
)


func TestGenerateWorkflowRunID(t *testing.T) {
	// ensure that the test data directory is deleted after the test
	defer os.RemoveAll(filepath.Join(config.GaleDataHome(), "testorg"))

	repo := &core.Repository{
		NameWithOwner: "testorg/testrepo",
	}

	// Test first workflow run ID generation
	wfRunID, err := idgen.GenerateWorkflowRunID(repo)
	if err != nil {
		t.Fatalf("Error generating workflow run ID: %v", err)
	}

	if wfRunID != "1" {
		t.Errorf("Expected first workflow run ID to be 1, got %s", wfRunID)
	}

	// Test second workflow run ID generation
	wfRunID, err = idgen.GenerateWorkflowRunID(repo)
	if err != nil {
		t.Fatalf("Error generating workflow run ID: %v", err)
	}

	if wfRunID != "2" {
		t.Errorf("Expected second workflow run ID to be 2, got %s", wfRunID)
	}

	repo2 := &core.Repository{
		NameWithOwner: "testorg/testrepo2",
	}

	// Test first workflow run ID generation
	wfRunID, err = idgen.GenerateWorkflowRunID(repo2)
	if err != nil {
		t.Fatalf("Error generating workflow run ID: %v", err)
	}

	if wfRunID != "1" {
		t.Errorf("Expected first workflow run ID to be 1, got %s", wfRunID)
	}
}

func TestGenerateJobRunID(t *testing.T) {
	// ensure that the test data directory is deleted after the test
	defer os.RemoveAll(filepath.Join(config.GaleDataHome(), "testorg"))

	repo := &core.Repository{
		NameWithOwner: "testorg/testrepo",
	}

	// Test first job run ID generation
	jobRunID, err := idgen.GenerateJobRunID(repo)
	if err != nil {
		t.Fatalf("Error generating job run ID: %v", err)
	}

	if jobRunID != "1" {
		t.Errorf("Expected first job run ID to be 1, got %s", jobRunID)
	}

	// Test second job run ID generation
	jobRunID, err = idgen.GenerateJobRunID(repo)
	if err != nil {
		t.Fatalf("Error generating job run ID: %v", err)
	}

	if jobRunID != "2" {
		t.Errorf("Expected second job run ID to be 2, got %s", jobRunID)
	}

	repo2 := &core.Repository{
		NameWithOwner: "testorg/testrepo2",
	}

	// Test first workflow run ID generation
	jobRunID, err = idgen.GenerateJobRunID(repo2)
	if err != nil {
		t.Fatalf("Error generating job run ID: %v", err)
	}

	if jobRunID != "1" {
		t.Errorf("Expected first job run ID to be 1, got %s", jobRunID)
	}
}
*/
