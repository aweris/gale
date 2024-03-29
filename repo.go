package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

// RepoInfo represents a repository information.
type RepoInfo struct {
	// Owner of the repository.
	Owner string

	// Name of the repository.
	Name string

	// NameWithOwner combined version of owner and name. Format: owner/name.
	NameWithOwner string

	// URL of the repository.
	URL string

	// Ref is the branch or tag ref that triggered the workflow
	Ref string

	// RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow.
	RefName string

	// RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither
	RefType string

	// SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	SHA string

	// ShortSHA is the short commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	ShortSHA string

	// IsRemote is true if the ref is a remote ref.
	IsRemote bool

	// Source is the directory containing the repository source.
	Source *Directory
}

func NewRepoInfo(ctx context.Context, source *Directory, repo, tag, branch string) (*RepoInfo, error) {
	// get the repository source from the options
	dir, err := getRepoSource(source, repo, tag, branch)
	if err != nil {
		return nil, err
	}

	// create a git container with the repository source to execute git commands
	container := dag.Container().From("alpine/git:latest").WithMountedDirectory("/src", dir).WithWorkdir("/src")

	// get the repository url
	url, err := getTrimmedOutput(ctx, container, "config", "--get", "remote.origin.url")
	if err != nil {
		return nil, err
	}

	// get the head commit short and long SHAs
	shortSHA, err := getTrimmedOutput(ctx, container, "rev-parse", "--short", "HEAD")
	if err != nil {
		return nil, err
	}

	sha, err := getTrimmedOutput(ctx, container, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	// get the ref and ref type
	ref, refType, err := getRefAndRefType(ctx, branch, tag, container, sha)
	if err != nil {
		return nil, err
	}

	// parse the github url to get the owner and repo name
	owner, repoName, err := parseGithubURL(url)
	if err != nil {
		return nil, err
	}

	return &RepoInfo{
		Owner:         owner,
		Name:          repoName,
		NameWithOwner: fmt.Sprintf("%s/%s", owner, repoName),
		URL:           url,
		Ref:           ref,
		RefName:       trimRefPrefix(ref),
		RefType:       refType,
		SHA:           sha,
		ShortSHA:      shortSHA,
		IsRemote:      source == nil,
		Source:        dir,
	}, nil
}

// workflows returns the workflows endpoint for the repository.
func (info *RepoInfo) workflows(dir string) *Workflows {
	return &Workflows{
		Repo:         info,
		WorkflowsDir: dir,
	}
}

// getRepoSource returns the repository source based on the options provided.
func getRepoSource(source *Directory, repo, tag, branch string) (*Directory, error) {
	if source != nil {
		return source, nil
	}

	if repo == "" {
		return nil, fmt.Errorf("either a repo or a source directory must be provided")
	}

	var (
		gitURL  = fmt.Sprintf("https://github.com/%s.git", repo)
		gitRepo = dag.Git(gitURL, GitOpts{KeepGitDir: true})
	)

	if tag != "" {
		return gitRepo.Tag(tag).Tree(), nil
	}

	if branch != "" {
		return gitRepo.Branch(branch).Tree(), nil
	}

	return nil, fmt.Errorf("when repo is provided, either a branch or a tag must be provided")
}

// getRefAndRefType returns the ref and ref type for given options.
func getRefAndRefType(
	ctx context.Context,
	tag string,
	branch string,
	container *Container,
	sha string,
) (string, string, error) {
	var (
		ref     string
		refType string
		err     error
	)

	// if branch or tag is provided, then repository cloned would be in detached head state. In that case, to work
	// around the issue, we're using given options to get the ref. If no branch or tag is provided, then we're using
	// the ref from the source code of the repository.

	if tag != "" {
		ref = fmt.Sprintf("refs/tags/%s", tag)
		refType = "tag"
		return ref, refType, nil
	}

	if branch != "" {
		ref = fmt.Sprintf("refs/heads/%s", branch)
		refType = "branch"
		return ref, refType, nil
	}

	ref, err = getRefFromSource(ctx, container, sha)
	if err != nil {
		return "", "", err
	}

	switch {
	case strings.HasPrefix(ref, "refs/tags/"):
		refType = "tag"
	case strings.HasPrefix(ref, "refs/heads/"):
		refType = "branch"
	default:
		return "", "", fmt.Errorf("unsupported ref type: %s", ref)
	}

	return ref, refType, nil
}

// getRefFromSource returns the ref for given head from the repository source.
func getRefFromSource(ctx context.Context, container *Container, head string) (string, error) {
	out, err := container.WithExec([]string{"show-ref"}).Stdout(ctx)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(strings.NewReader(out))

	found := ""
	for scanner.Scan() {
		ref := scanner.Text()

		parts := strings.Fields(ref)

		if len(parts) < 2 {
			continue
		}

		ref = strings.TrimSpace(parts[0])

		if ref == head {
			found = strings.TrimSpace(parts[1])
			break
		}
	}

	if found == "" {
		return "", fmt.Errorf("no ref found for %s", head)
	}

	return found, nil
}

// parseGithubURL parses the github url and returns the owner and repo.
func parseGithubURL(url string) (string, string, error) {
	var owner, repo string

	trimGitHubURL := func(url, prefix string) (string, string) {
		trimmed := strings.TrimPrefix(url, prefix)

		parts := strings.Split(trimmed, "/")

		return parts[0], strings.TrimSuffix(parts[1], ".git")
	}

	switch {
	case strings.HasPrefix(url, "git@github.com:"):
		owner, repo = trimGitHubURL(url, "git@github.com:")
	case strings.HasPrefix(url, "https://github.com/"):
		owner, repo = trimGitHubURL(url, "https://github.com/")
	default:
		return "", "", fmt.Errorf("unsupported GitHub URL format: %s", url)
	}

	if owner == "" || repo == "" {
		return "", "", fmt.Errorf("invalid GitHub URL: %s", url)
	}

	return owner, repo, nil
}

// trimRefPrefix trims the prefix from the ref.
func trimRefPrefix(ref string) string {
	switch {
	case strings.HasPrefix(ref, "refs/tags/"):
		return strings.TrimPrefix(ref, "refs/tags/")
	case strings.HasPrefix(ref, "refs/heads/"):
		return strings.TrimPrefix(ref, "refs/heads/")
	default:
		return ref
	}
}

// getTrimmedOutput returns the trimmed output of the command executed in the container.
func getTrimmedOutput(ctx context.Context, container *Container, args ...string) (string, error) {
	out, err := container.WithExec(args).Stdout(ctx)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}
