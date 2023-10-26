package main

import (
	"bufio"
	"context"
	"fmt"
	"strings"
)

type Repo struct{}

// RepoInfo represents a repository information.
type RepoInfo struct {
	Owner         string // Owner of the repository.
	Name          string // Name of the repository.
	NameWithOwner string // NameWithOwner combined version of owner and name. Format: owner/name.
	URL           string // URL of the repository.
	Ref           string // Ref is the branch or tag ref that triggered the workflow
	RefName       string // RefName is the short name (without refs/heads/ prefix) of the branch or tag ref that triggered the workflow.
	RefType       string // RefType is the type of ref that triggered the workflow. Possible values are branch, tag, or empty, if neither
	SHA           string // SHA is the commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	ShortSHA      string // ShortSHA is the short commit SHA that triggered the workflow. The value of this commit SHA depends on the event that
	IsRemote      bool   // IsRemote is true if the ref is a remote ref.
}

// TODO: follow up
// this method is separate from the RepoInfo struct because we're not able to return *Directory as part of RepoInfo.
// Until it is fixed, we're returning *Directory from this method.

func (_ *Repo) Source(
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
) (*Directory, error) {
	return getRepoSource(source, repo, tag, branch)
}

func (_ *Repo) Info(
	ctx context.Context,
	// The directory containing the repository source. If source is provided, rest of the options are ignored.
	source Optional[*Directory],
	// The name of the repository. Format: owner/name.
	repo Optional[string],
	// Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	tag Optional[string],
	// Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
	branch Optional[string],
) (*RepoInfo, error) {
	// get the repository source from the options
	dir, err := getRepoSource(source, repo, tag, branch)
	if err != nil {
		return nil, err
	}

	// create a git container with the repository source to execute git commands
	container := gitContainer(dir)

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

	_, isLocal := source.Get()

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
		IsRemote:      !isLocal,
	}, nil
}

// Workdir returns the runner workdir for the repository.
func (ri *RepoInfo) Workdir() string {
	return fmt.Sprintf("/home/runner/work/%s/%s", ri.Name, ri.Name)
}

// Configure configures the container with the repository information.
func (ri *RepoInfo) Configure(_ context.Context, c *Container) (*Container, error) {
	return c.WithEnvVariable("GH_REPO", ri.NameWithOwner).
		WithEnvVariable("GITHUB_REPOSITORY", ri.NameWithOwner).
		WithEnvVariable("GITHUB_REPOSITORY_OWNER", ri.Owner).
		WithEnvVariable("GITHUB_REPOSITORY_URL", ri.URL).
		WithEnvVariable("GITHUB_REF", ri.Ref).
		WithEnvVariable("GITHUB_REF_NAME", ri.RefName).
		WithEnvVariable("GITHUB_REF_TYPE", ri.RefType).
		WithEnvVariable("GITHUB_SHA", ri.SHA), nil
}

// getRepoSource returns the repository source based on the options provided.
func getRepoSource(sourceOpt Optional[*Directory], repoOpt, tagOpt, branchOpt Optional[string]) (*Directory, error) {
	if source, ok := sourceOpt.Get(); ok {
		return source, nil
	}

	repo, ok := repoOpt.Get()
	if !ok {
		return nil, fmt.Errorf("either a repo or a source directory must be provided")
	}

	var (
		gitURL  = fmt.Sprintf("https://github.com/%s.git", repo)
		gitRepo = dag.Git(gitURL, GitOpts{KeepGitDir: true})
	)

	if tag, ok := tagOpt.Get(); ok {
		return gitRepo.Tag(tag).Tree(), nil
	}

	if branch, ok := branchOpt.Get(); ok {
		return gitRepo.Branch(branch).Tree(), nil
	}

	return nil, fmt.Errorf("when repo is provided, either a branch or a tag must be provided")
}

// getRefAndRefType returns the ref and ref type for given options.
func getRefAndRefType(
	ctx context.Context,
	tagOpt Optional[string],
	branchOpt Optional[string],
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

	if tag, ok := tagOpt.Get(); ok {
		ref = fmt.Sprintf("refs/tags/%s", tag)
		refType = "tag"
		return ref, refType, nil
	}

	if branch, ok := branchOpt.Get(); ok {
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
