package images

import (
	"runtime"
	"strings"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
)

// GoBase returns a container with the base image for the go
func GoBase() *dagger.Container {
	return config.Client().Container().From("golang:" + strings.TrimPrefix(runtime.Version(), "go"))
}
