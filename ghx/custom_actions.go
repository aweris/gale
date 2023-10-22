package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	"github.com/aweris/gale/common/fs"
	"github.com/aweris/gale/common/log"
	"github.com/aweris/gale/ghx/core"
)

// LoadActionFromSource loads an action from given source to the target directory. If the source is a local action,
// the target directory will be the same as the source. If the source is a remote action, the action will be downloaded
// to the target directory using the source as the reference(e.g. {target}/{owner}/{repo}/{path}@{ref}).
func LoadActionFromSource(ctx context.Context, client *dagger.Client, source, targetDir string) (*core.CustomAction, error) {
	var target string

	repo, path, ref, err := parseRepoRef(source)
	if err != nil {
		return nil, err
	}

	// no need to load action if it is a local action
	if isLocalAction(source) {
		target = source
	} else {
		target = filepath.Join(targetDir, source)

		// ensure action exists locally -- FIXME: source just passed for logging purposes, should be refactored
		if err := ensureActionExistsLocally(source, repo, ref, target); err != nil {
			return nil, err
		}
	}

	dir := client.Host().Directory(target)

	if path != "" {
		dir = dir.Directory(path) // if path is not empty, read the action from the path
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
func ensureActionExistsLocally(source, repo, ref, target string) error {
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

	url := fmt.Sprintf("https://github.com/%s.git", repo)

	// Clone the repository into the target directory using go-git
	r, err := git.PlainClone(target, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone action repository: %w", err)
	}

	// Resolve the revision (be it a branch, tag or commit hash)
	hash, err := r.ResolveRevision(plumbing.Revision(ref))
	if err != nil {
		return fmt.Errorf("failed to resolve revision: %w", err)
	}

	// Checkout to the commit
	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{Hash: *hash})
	if err != nil {
		return fmt.Errorf("failed to checkout: %w", err)
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
