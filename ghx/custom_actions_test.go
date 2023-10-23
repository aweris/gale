package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"dagger.io/dagger"
)

func TestCustomActionManager_GetCustomAction(t *testing.T) {
	dir, err := os.MkdirTemp("", "ghx-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	ctx := context.Background()

	client, err := dagger.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("download missing action", func(t *testing.T) {
		ca, err := LoadActionFromSource(ctx, client, "actions/checkout@v2", filepath.Join(dir, "actions"))
		if err != nil {
			t.Fatal(err)
		}

		if ca.Path != filepath.Join(dir, "actions", "actions/checkout@v2") {
			t.Fatal("action dir is different than expected")
		}

		if _, err := os.Stat(filepath.Join(dir, "actions", "actions/checkout@v2")); err != nil {
			t.Fatal("action dir not exported")
		}

		if ca.Meta.Name != "Checkout" {
			t.Fatalf("action name mismatch expected: checkout, actual: %s", ca.Meta.Name)
		}
	})

	t.Run("return existing action", func(t *testing.T) {
		target := filepath.Join(dir, "actions", "some/action@v1")

		// create the action directory
		if err := os.MkdirAll(target, 0700); err != nil {
			t.Fatal(err)
		}

		// create the action metadata file
		err := os.WriteFile(filepath.Join(target, "action.yml"), []byte("name: some-action"), 0600)
		if err != nil {
			t.Fatal(err)
		}

		ca, err := LoadActionFromSource(ctx, client, "some/action@v1", filepath.Join(dir, "actions"))
		if err != nil {
			t.Fatal(err)
		}

		if ca.Path != target {
			t.Fatal("action dir is different than expected")
		}

		if ca.Meta.Name != "some-action" {
			t.Fatal("action name mismatch")
		}
	})
}
