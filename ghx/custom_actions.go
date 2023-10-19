package main

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/core"
)

// LoadActionFromSource loads an action from given source to the target directory. If the source is a local action,
// the target directory will be the same as the source. If the source is a remote action, the action will be downloaded
// to the target directory using the source as the reference(e.g. {target}/{owner}/{repo}/{path}@{ref}).
func LoadActionFromSource(ctx context.Context, client *dagger.Client, source, targetDir string) (*core.CustomAction, error) {
	var target string

	// no need to load action if it is a local action
	if isLocalAction(source) {
		target = source
	} else {
		target = filepath.Join(targetDir, source)

		// ensure action exists locally
		if err := ensureActionExistsLocally(ctx, client, source, target); err != nil {
			return nil, err
		}
	}

	dir, err := getActionDirectory(client, target)
	if err != nil {
		return nil, err
	}

	meta, err := getCustomActionMeta(ctx, dir)
	if err != nil {
		return nil, err
	}

	return &core.CustomAction{Meta: meta, Path: target, Dir: dir}, nil
}

// isLocalAction checks if the given source is a local action
func isLocalAction(source string) bool {
	return strings.HasPrefix(source, "./") || filepath.IsAbs(source) || strings.HasPrefix(source, "/")
}

// ensureActionExistsLocally ensures that the action exists locally. If the action does not exist locally, it will be
// downloaded from the source to the target directory.
func ensureActionExistsLocally(ctx context.Context, client *dagger.Client, source, target string) error {
	// check if action exists locally
	exist, err := fs.Exists(target)
	if err != nil {
		return fmt.Errorf("failed to check if action exists locally: %w", err)
	}

	// do nothing if target path already exists
	if exist {
		log.Debugf("action already exists locally", "source", source, "target", target)
		return nil
	}

	log.Debugf("action does not exist locally, downloading...", "source", source, "target", target)

	dir, err := getActionDirectory(client, source)
	if err != nil {
		return err
	}

	// export the action to the target directory
	_, err = dir.Export(ctx, target)
	if err != nil {
		return err
	}

	return nil
}

// getCustomActionMeta returns the meta information about the custom action from the action directory.
func getCustomActionMeta(ctx context.Context, actionDir *dagger.Directory) (core.CustomActionMeta, error) {
	var meta core.CustomActionMeta

	file, err := findActionMetadataFileName(ctx, actionDir)
	if err != nil {
		return meta, err
	}

	content, err := actionDir.File(file).Contents(ctx)
	if err != nil {
		return meta, err
	}

	err = yaml.Unmarshal([]byte(content), &meta)
	if err != nil {
		return meta, err
	}

	return meta, nil
}

// getActionDirectory returns the directory of the action from given source.
func getActionDirectory(client *dagger.Client, source string) (*dagger.Directory, error) {
	// if path is relative, use the host to resolve the path
	if isLocalAction(source) {
		return client.Host().Directory(source), nil
	}

	// if path is not a relative path, it must be a remote repository in the format "{owner}/{repo}/{path}@{ref}"
	// if {path} is not present in the input string, an empty string is returned for the path component.
	actionRepo, actionPath, actionRef, err := parseRepoRef(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo ref %s: %v", source, err)
	}

	// TODO: handle enterprise github instances as well
	gitRepo := client.Git(path.Join("github.com", actionRepo))

	var gitRef *dagger.GitRef

	refType, err := determineRefTypeFromRepo(actionRepo, actionRef)
	if err != nil {
		return nil, err
	}

	switch refType {
	case core.RefTypeBranch:
		gitRef = gitRepo.Branch(actionRef)
	case core.RefTypeTag:
		gitRef = gitRepo.Tag(actionRef)
	case core.RefTypeCommit:
		gitRef = gitRepo.Commit(actionRef)
	}

	dir := gitRef.Tree()

	if actionPath != "" {
		dir = dir.Directory(actionPath)
	}

	return dir, nil
}

// findActionMetadataFileName finds the action.yml or action.yaml file in the root of the action directory.
func findActionMetadataFileName(ctx context.Context, dir *dagger.Directory) (string, error) {
	// list all entries in the root of the action directory
	entries, entriesErr := dir.Entries(ctx)
	if entriesErr != nil {
		return "", fmt.Errorf("failed to list entries for: %v", entriesErr)
	}

	file := ""

	// find action.yml or action.yaml exists in the root of the action repo
	for _, entry := range entries {
		if entry == "action.yml" || entry == "action.yaml" {
			file = entry
			break
		}
	}

	// if action.yml or action.yaml does not exist, return an error
	if file == "" {
		return "", fmt.Errorf("action.yml or action.yaml not found in the root of the action directory")
	}

	return file, nil
}

// parseRepoRef parses a string in the format "{owner}/{repo}/{path}@{ref}" and returns the parsed components.
// If {path} is not present in the input string, an empty string is returned for the path component.
func parseRepoRef(input string) (repo string, path string, ref string, err error) {
	regex := regexp.MustCompile(`^([^/]+)/([^/@]+)(?:/([^@]+))?@(.+)$`)
	matches := regex.FindStringSubmatch(input)

	if len(matches) == 0 {
		err = fmt.Errorf("invalid input format: %q", input)
		return
	}

	repo = strings.Join([]string{matches[1], matches[2]}, "/")
	path = matches[3]
	ref = matches[4]

	return
}

// determineRefTypeFromRepo determines the type of ref from given repository and ref. The ref can be either a branch or a tag.
// If the ref is not a branch, a tag or a commit, it will return RefTypeUnknown.
//
// The method will use GitHub API to determine the type of ref. If the ref does not exist on remote, it will
// return RefTypeUnknown.
func determineRefTypeFromRepo(repo, ref string) (core.RefType, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return core.RefTypeUnknown, fmt.Errorf("failed to create github client: %w\n", err)
	}

	var dummy interface{}

	err = client.Get(fmt.Sprintf("repos/%s/git/ref/heads/%s", repo, ref), &dummy)
	if err == nil {
		return core.RefTypeBranch, nil
	}

	err = client.Get(fmt.Sprintf("repos/%s/git/ref/tags/%s", repo, ref), &dummy)
	if err == nil {
		return core.RefTypeTag, nil
	}

	err = client.Get(fmt.Sprintf("repos/%s/git/commits/%s", repo, ref), &dummy)
	if err == nil {
		return core.RefTypeCommit, nil
	}

	log.Warn(fmt.Sprintf("%s repo does not have tag, branch or commit ref for %s", repo, ref))

	return core.RefTypeUnknown, fmt.Errorf("failed to determine ref type for %s@%s\n", repo, ref)
}
