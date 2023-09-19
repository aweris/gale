package gctx

import (
	"context"
	"os"
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
	CacheVol     *data.CacheVolume     `json:"-"`
	Source       *dagger.Directory     `json:"-"`
	Info         core.Repository       `json:"info" container_env:"true"`
	Ref          core.RepositoryGitRef `json:"ref" container_env:"true"`
	WorkflowsDir string                `json:"workflows_dir" env:"GALE_WORKFLOWS_DIR" envDefault:".github/workflows" container_env:"true"`
}

// LoadRepo initializes the context with the specified or current repository's default branch if no options are provided.
func (c *Context) LoadRepo(repo string, opts ...LoadRepoOpts) error {
	// load repo context from env
	rc, err := NewContextFromEnv[RepoContext]()
	if err != nil {
		return err
	}

	// in container mode, we don't need to load the repository information from the host, we can use the one from the
	// environment.
	if !c.isContainer {
		rc.Info, err = core.GetRepository(repo)
		if err != nil {
			return err
		}
	}

	// load repository source - it doesn't matter if we're in container mode or not, we need to load the repository

	opt := LoadRepoOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// load repository source
	rc.Source = getRepository(repo, rc.Info, opt)

	// cache volume only used non container mode. Instead of leaving it nil, we create a new one in container mode kas
	// well to avoid nil pointer errors. There is no harm in doing so since we're supporting Dagger in Dagger mode.
	rc.CacheVol = data.NewCacheVolume(rc.Info)

	// load rest of the repository context since we have the repository information and the source. Same here, in
	// container mode, we don't need to load the repository ref from the host, we can use the one from the environment.
	if !c.isContainer {
		rc.Ref, err = getRepositoryRef(c.Context, repo, opt, rc.Source)
		if err != nil {
			return err
		}

		rc.WorkflowsDir = opt.WorkflowsDir

		c.Github.setRepo(rc.Info, rc.Ref)
	}

	c.Repo = rc

	return nil
}

// getRepository returns a dagger directory for the specified repository and options. If repo and options are empty,
// the current git repository will be used as it is. If repo or any of the options are provided, the repository will
// be cloned from the specified url and the specified tag or branch will be checked out.
func getRepository(repo string, info core.Repository, opt LoadRepoOpts) (source *dagger.Directory) {
	//  if repo and options are empty, use the current repository
	if repo == "" && opt.Tag == "" && opt.Branch == "" {
		return config.Client().Host().Directory(".")
	}

	switch {
	case opt.Tag != "":
		source = config.Client().Git(info.URL, dagger.GitOpts{KeepGitDir: true}).Tag(opt.Tag).Tree()
	case opt.Branch != "":
		source = config.Client().Git(info.URL, dagger.GitOpts{KeepGitDir: true}).Branch(opt.Branch).Tree()
	default:
		source = config.Client().Git(info.URL, dagger.GitOpts{KeepGitDir: true}).Branch(info.DefaultBranchRef.Name).Tree()
	}

	return source
}

// getRepositoryRef gets the repository ref from the specified repository and options.
func getRepositoryRef(ctx context.Context, repo string, opt LoadRepoOpts, source *dagger.Directory) (ref core.RepositoryGitRef, err error) {
	path := "."

	// if repo is not a local directory, we need to export it to a temp directory and use that as the path for the
	// getting the ref from the git repository directly.
	//
	// for this conditional:
	//  - if repo is not empty, it means that the user has provided a remote repo
	//  - if repo is empty, but tag or branch is provided, it means that the user wants to use the current repository
	//  with the specified tag or branch
	//
	// So for both cases, we're using dagger git to clone the repository to a directory, no local directory is
	// provided.
	if repo != "" || opt.Tag != "" && opt.Branch != "" {
		dir, err := os.MkdirTemp("/tmp", strings.ReplaceAll(repo, "/", "-"))
		if err != nil {
			return ref, err
		}
		defer os.RemoveAll(dir)

		_, err = source.Export(ctx, dir)
		if err != nil {
			return ref, err
		}

		path = dir
	}

	return core.GetRepositoryRefFromDir(path, opt.Tag, opt.Branch)
}

// LoadCurrentRepo initializes the context with the repository information from the current working directory,
// using the specified options or current repository state if none are provided.
func (c *Context) LoadCurrentRepo(opts ...LoadRepoOpts) error {
	return c.LoadRepo("", opts...)
}

var _ helpers.WithContainerFuncHook = new(RepoContext)

func (c *RepoContext) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		return container.With(WithContainerEnv(config.Client(), c))
	}
}

func (c *Context) LoadWorkflows() (map[string]core.Workflow, error) {
	dir := c.Repo.Source.Directory(c.Repo.WorkflowsDir)

	entries, err := dir.Entries(c.Context)
	if err != nil {
		return nil, err
	}

	workflows := make(map[string]core.Workflow)

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)

			workflow, err := loadWorkflow(c.Context, filepath.Join(c.Repo.WorkflowsDir, entry), file)
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
