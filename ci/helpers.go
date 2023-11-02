package main

import (
	"path/filepath"
	"runtime"
)

// root returns the root directory of the project.
func root() string {
	// get location of current file
	_, current, _, _ := runtime.Caller(0)

	return filepath.Join(filepath.Dir(current), "..")
}
