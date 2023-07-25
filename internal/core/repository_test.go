package core_test

import (
	"context"
	"dagger.io/dagger"
	"github.com/aweris/gale/internal/config"
	"os"
	"testing"

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

	entries, err := repo.Dir.Entries(ctx)
	if err != nil {
		t.Fatalf("Failed to get directory entries: %s", err)
	}

	if len(entries) == 0 {
		t.Fatalf("Expected directory entries to be more than 0 but got %d", len(entries))
	}

	data, err := repo.Dir.File("testdata/content.txt").Contents(ctx)
	if err != nil {
		t.Fatalf("Failed to get file contents: %s", err)
	}

	if string(data) != "Hello World" {
		t.Fatalf("Expected file contents to be Hello World but got %s", string(data))
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

	entries, err := repo.Dir.Entries(ctx)
	if err != nil {
		t.Fatalf("Failed to get directory entries: %s", err)
	}

	if len(entries) == 0 {
		t.Fatalf("Expected directory entries to be more than 0 but got %d", len(entries))
	}
}
