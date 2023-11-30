package main

import (
	"fmt"

	"github.com/google/uuid"
)

// default entry point for internal submodules. The intention of this entry point is keep the module clean and
// consistent. This entrypoint is not intended to be used by external modules.
var internal Internal

type Internal struct{}

func (_ *Internal) runner() *Runner {
	return &Runner{}
}

func (_ *Internal) repo() *Repo {
	return &Repo{}
}

func (_ *Internal) workflows() *Workflows {
	return &Workflows{}
}

func (_ *Internal) context() *RunContext {
	var (
		rid  = uuid.New().String()
		data = dag.CacheVolume(fmt.Sprintf("ghx-run-%s", rid))
	)

	return &RunContext{RunID: rid, SharedData: data}
}
