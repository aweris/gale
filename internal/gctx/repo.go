package gctx

import (
	"context"
	"path/filepath"
	"strings"

	"dagger.io/dagger"
	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/dagger/helpers"
	"github.com/aweris/gale/pkg/data"
)

type LoadRepoOpts struct {
	Branch       string
	Tag          string
	WorkflowsDir string
}

type RepoContext struct {
	Repository   *core.Repository         `json:"-"`
	CacheVol     *data.CacheVolume        `json:"-"`
	WorkflowsDir string                   `json:"workflows_dir" env:"GALE_WORKFLOWS_DIR" envDefault:".github/workflows" container_env:"true"`
	Workflows    map[string]core.Workflow `json:"-"`
}

// LoadRepo initializes the context with the specified or current repository's default branch if no options are provided.
func (c *Context) LoadRepo(repo string, opts ...LoadRepoOpts) error {
	rc, err := NewContextFromEnv[RepoContext]()
	if err != nil {
		return err
	}

	opt := LoadRepoOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	r, err := core.GetRepository(repo, core.GetRepositoryOpts{Branch: opt.Branch, Tag: opt.Tag})
	if err != nil {
		return err
	}

	rc.Repository = r

	// cache volume only used non container mode. Instead of leaving it nil, we create a new one in container mode kas
	// well to avoid nil pointer errors. There is no harm in doing so since we're supporting Dagger in Dagger mode.
	rc.CacheVol = data.NewCacheVolume(r)

	// if it's not in container mode, workflows dir should be set from the options
	if !c.isContainer {
		rc.WorkflowsDir = opt.WorkflowsDir
	}

	// load workflows
	workflows, err := loadWorkflows(c.Context, r, rc.WorkflowsDir)
	if err != nil {
		return err
	}

	rc.Workflows = workflows

	c.Repo = rc

	return nil
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo(opts ...LoadRepoOpts) error {
	return c.LoadRepo("", opts...)
}

var _ helpers.WithContainerFuncHook = new(RepoContext)

func (c RepoContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.With(WithContainerEnv(config.Client(), c))
	}
}

func loadWorkflows(ctx context.Context, repo *core.Repository, path string) (map[string]core.Workflow, error) {
	dir := repo.GitRef.Dir.Directory(path)

	entries, err := dir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make(map[string]core.Workflow)

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)

			workflow, err := loadWorkflow(ctx, filepath.Join(path, entry), file)
			if err != nil {
				return nil, err
			}

			workflows[workflow.Name] = workflow
		}
	}

	return workflows, nil
}

// loadWorkflow loads a workflow from a file. If the workflow name is not provided, the relative path to the workflow
// file will be used as the workflow name.
func loadWorkflow(ctx context.Context, path string, file *dagger.File) (workflow core.Workflow, err error) {
	content, err := file.Contents(ctx)
	if err != nil {
		return workflow, err
	}

	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return workflow, err
	}

	// set workflow path
	workflow.Path = path

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = path
	}

	// update job ID and names
	for id, job := range workflow.Jobs {
		job.ID = id

		if job.Name == "" {
			job.Name = id
		}

		workflow.Jobs[id] = job
	}

	return workflow, nil
}
