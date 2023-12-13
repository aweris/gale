package main

import (
	"fmt"
	"strings"
)

// dagger returns a dagger container with the specified dagger version. If no version is specified, the latest version
// will be used.
func dagger(daggerVersion Optional[string]) *Container {
	version, ok := daggerVersion.Get()

	if ok {
		// remove the leading v if it exists
		version = strings.TrimPrefix(version, "v")

		// prefix the version with DAGGER_VERSION=
		version = fmt.Sprintf("DAGGER_VERSION=%s", version)
	}

	return dag.Container().From("alpine:latest").
		WithExec([]string{"apk", "add", "curl"}).
		WithExec([]string{"sh", "-c", fmt.Sprintf("curl -L https://dl.dagger.io/dagger/install.sh | %s sh", version)}).
		WithEntrypoint([]string{"/bin/dagger"})
}
