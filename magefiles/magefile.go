//go:build mage

package main

import (
	"context"
	"os"

	"github.com/magefile/mage/sh"
)

// Run is a temporary function to execute the code
func Run(ctx context.Context) error {
	_, err := sh.Exec(map[string]string{}, os.Stdout, os.Stderr, "go", "run", "./cmd/gale")
	if err != nil {
		return err
	}

	return nil
}
