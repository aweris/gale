package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// default entry point for internal submodules. The intention of this entry point is keep the module clean and
// consistent. This entrypoint is not intended to be used by external modules.
var internal Internal

type Internal struct{}

func (_ *Internal) runner(base *Container, info *RepoInfo) *Runner {
	return &Runner{BaseContainer: base, Repo: info}
}

func (_ *Internal) repo() *Repo {
	return &Repo{}
}

func (_ *Internal) context(opts *WorkflowRunOpts) *RunContext {
	var (
		rid  = uuid.New().String()
		data = dag.CacheVolume(fmt.Sprintf("ghx-run-%s", rid))
	)

	return &RunContext{RunID: rid, Opts: opts, SharedData: data}
}

// getWorkflow returns the workflow with the given options. IF workflowFile is provided, it will be used. Otherwise,
// workflow will be loaded from the repository source with the given options.
func (_ *Internal) getWorkflow(ctx context.Context, info *RepoInfo, file *File, workflow string, dir string) (*Workflow, error) {
	// FIXME: when dagger supports accepting common input/output types like Custom structs or interfaces from different
	//  modules, we can refactor this to accept a common Workflow type instead of two different options.

	if file != nil {
		return info.workflows().loadWorkflow(ctx, "", file)
	}

	if workflow == "" {
		return nil, fmt.Errorf("workflow or workflow file must be provided")
	}

	return info.workflows().Get(ctx, workflow, dir)
}
