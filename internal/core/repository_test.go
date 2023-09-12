package core_test

import (
	"context"
	"os"
	"testing"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
)

func TestGetCurrentRepository(t *testing.T) {
	ctx := context.Background()

	// set dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	config.SetClient(client)

	// get current repository with current directory. This will load tests directory as repository. It's ok for testing.
	repo, err := core.GetCurrentRepository()
	if err != nil {
		t.Fatalf("Failed to get current repository: %s", err)
	}

	if repo.Name != "gale" {
		t.Fatalf("Expected repository name to be gale but got %s", repo.Name)
	}

	if repo.NameWithOwner != "aweris/gale" {
		t.Fatalf("Expected repository name with owner to be aweris/gale but got %s", repo.NameWithOwner)
	}

	entries, err := repo.GitRef.Dir.Entries(ctx)
	if err != nil {
		t.Fatalf("Failed to get directory entries: %s", err)
	}

	if len(entries) == 0 {
		t.Fatalf("Expected directory entries to be more than 0 but got %d", len(entries))
	}

	data, err := repo.GitRef.Dir.File("testdata/content.txt").Contents(ctx)
	if err != nil {
		t.Fatalf("Failed to get file contents: %s", err)
	}

	if data != "Hello World" {
		t.Fatalf("Expected file contents to be Hello World but got %s", data)
	}
}

func TestGetRepository(t *testing.T) {
	ctx := context.Background()

	// set dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	config.SetClient(client)

	// get repository with default branch
	repo, err := core.GetRepository("aweris/gale")
	if err != nil {
		t.Fatalf("Failed to get current repository: %s", err)
	}

	if repo.Name != "gale" {
		t.Fatalf("Expected repository name to be gale but got %s", repo.Name)
	}

	if repo.NameWithOwner != "aweris/gale" {
		t.Fatalf("Expected repository name with owner to be aweris/gale but got %s", repo.NameWithOwner)
	}

	entries, err := repo.GitRef.Dir.Entries(ctx)
	if err != nil {
		t.Fatalf("Failed to get directory entries: %s", err)
	}

	if len(entries) == 0 {
		t.Fatalf("Expected directory entries to be more than 0 but got %d", len(entries))
	}
}

/*
FIXME: This test is failing because of the LoadWorkflows function moved to gctx package. We need to refactor this test

func TestLoadWorkflows(t *testing.T) {
	ctx := context.Background()

	// set dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	config.SetClient(client)

	// get current repository with current directory. This will load tests directory as repository. It's ok for testing.
	repo, err := core.GetCurrentRepository()
	if err != nil {
		t.Fatalf("Failed to get current repository: %s", err)
	}

	workflows, err := repo.LoadWorkflows(ctx, core.RepositoryLoadWorkflowOpts{WorkflowsDir: "testdata/workflows"})
	if err != nil {
		t.Fatalf("Failed to load workflows: %s", err)
	}

	if len(workflows) != 2 {
		t.Fatalf("Expected workflows to be 2 but got %d", len(workflows))
	}

	if _, ok := workflows["test-with-name"]; !ok {
		t.Fatalf("Expected workflow example-golangci-lint to be loaded but not found")
	}

	if _, ok := workflows["testdata/workflows/test-workflow-without-name.yml"]; !ok {
		t.Fatalf("Expected workflow example-golangci-lint to be loaded but not found")
	}
}

*/
