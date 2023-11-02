package main

import (
	"os"
	"path/filepath"
)

type Daggerverse struct{}

// Syncs all daggerverse modules with given dagger version and returns the daggerverse directory with updated modules.
func (m *Daggerverse) Sync(
	// The dagger version to use. If not specified, the latest version will be used.
	daggerVersion Optional[string],
) (*Directory, error) {
	container := dagger(daggerVersion).
		WithMountedDirectory("/src", dag.Host().Directory(root())).
		WithWorkdir("/src")

	modules, err := m.list()
	if err != nil {
		return nil, err
	}

	for _, module := range modules {
		container = container.WithExec([]string{"mod", "-m", module, "sync"}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true})
	}

	return container.Directory("/src/daggerverse"), nil
}

func (m *Daggerverse) list() ([]string, error) {
	var modules []string

	// Walk the start directory recursively
	err := filepath.Walk(filepath.Join(root(), "daggerverse"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directory files
		if !info.IsDir() {
			return nil
		}

		// Check if the directory contains a file named "dagger.json"
		daggerJSONPath := filepath.Join(path, "dagger.json")
		if _, err := os.Stat(daggerJSONPath); err == nil {
			// Get the relative path to the directory from the start directory
			relPath, err := filepath.Rel(root(), path)
			if err != nil {
				return err
			}
			modules = append(modules, relPath)
		}
		return nil
	})

	return modules, err
}
