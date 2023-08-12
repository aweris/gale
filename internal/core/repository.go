package core

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/cli/go-gh/v2"

	"github.com/aweris/gale/internal/config"
)

// Repository represents a GitHub repository
type Repository struct {
	ID               string
	Name             string
	NameWithOwner    string
	URL              string
	Owner            RepositoryOwner
	DefaultBranchRef RepositoryBranchRef
	GitRef           *RepositoryGitRef
}

// RepositoryOwner represents a GitHub repository owner
type RepositoryOwner struct {
	ID    string
	Login string
}

// RepositoryBranchRef represents a GitHub repository branch ref
type RepositoryBranchRef struct {
	Name string
}

// GetRepositoryOpts represents the options for refs used to get a repository. Only one of the
// options can be set.
//
// If none of the options are set, the default branch will be used. Default branch or remote repositories is configured
// in the GitHub repository settings. For local repositories, it's the branch that is currently checked out.
//
// If multiple options are set, the precedence is as follows: tag, branch.
type GetRepositoryOpts struct {
	Branch string
	Tag    string
}

// GetCurrentRepository returns current repository information. This is a wrapper around GetRepository with empty name.
func GetCurrentRepository() (*Repository, error) {
	return GetRepository("")
}

// GetRepository returns repository information. If name is empty, the current repository will be used.
func GetRepository(name string, opts ...GetRepositoryOpts) (*Repository, error) {
	var repo Repository

	stdout, stderr, err := gh.Exec("repo", "view", name, "--json", "id,name,owner,nameWithOwner,url,defaultBranchRef")
	if err != nil {
		return nil, fmt.Errorf("failed to get current repository: %w stderr: %s", err, stderr.String())
	}

	err = json.Unmarshal(stdout.Bytes(), &repo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal current repository: %s err: %w", stdout.String(), err)
	}

	opt := GetRepositoryOpts{}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// load repo tree based on the options precedence
	switch {
	case opt.Tag != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeTag, opt.Tag)
		if err != nil {
			return nil, err
		}
	case opt.Branch != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeBranch, opt.Branch)
		if err != nil {
			return nil, err
		}
	case name != "":
		repo.GitRef, err = GetRepositoryGitRef(context.Background(), repo.URL, RefTypeBranch, repo.DefaultBranchRef.Name)
		if err != nil {
			return nil, err
		}
	default:
		// TODO: current directory could be a subdirectory of the repository. Should we handle this case?
		repo.GitRef, err = GetRepositoryRefFromDir(context.Background(), config.Client().Host().Directory("."))
		if err != nil {
			return nil, err
		}
	}

	return &repo, nil
}

// RepositoryLoadWorkflowOpts represents the options for loading workflows
type RepositoryLoadWorkflowOpts struct {
	WorkflowsDir string // WorkflowsDir is the path to the workflow file. If empty, default path .github/workflows will be used.
}

func (r *Repository) LoadWorkflows(ctx context.Context, opts ...RepositoryLoadWorkflowOpts) (map[string]*Workflow, error) {
	path := ".github/workflows"

	if len(opts) > 0 {
		if opts[0].WorkflowsDir != "" {
			path = opts[0].WorkflowsDir
		}
	}

	dir := r.GitRef.Dir.Directory(path)

	entries, err := dir.Entries(ctx)
	if err != nil {
		return nil, err
	}

	workflows := make(map[string]*Workflow)

	for _, entry := range entries {
		// load only .yaml and .yml files
		if strings.HasSuffix(entry, ".yaml") || strings.HasSuffix(entry, ".yml") {
			file := dir.File(entry)

			workflow, err := r.loadWorkflow(ctx, filepath.Join(path, entry), file)
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
func (r *Repository) loadWorkflow(ctx context.Context, path string, file *dagger.File) (*Workflow, error) {
	content, err := file.Contents(ctx)
	if err != nil {
		return nil, err
	}

	var workflow Workflow

	if err := yaml.Unmarshal([]byte(content), &workflow); err != nil {
		return nil, err
	}

	workflow.Path = path

	// if the workflow name is not provided, use the relative path to the workflow file.
	if workflow.Name == "" {
		workflow.Name = path
	}

	return &workflow, nil
}
