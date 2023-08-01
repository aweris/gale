package actions

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"dagger.io/dagger"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/core"
	"github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/tools/ghx/log"
)

// LoadActionFromSource loads an action from given source. Source can be a local directory or a remote repository.
func LoadActionFromSource(ctx context.Context, source string) (*core.CustomAction, error) {
	var target string

	// no need to load action if it is a local action
	if isLocalAction(source) {
		target = source
	} else {
		target = filepath.Join(config.GhxActionsDir(), source)

		// ensure action exists locally
		if err := ensureActionExistsLocally(ctx, source, target); err != nil {
			return nil, err
		}
	}

	dir, err := getActionDirectory(target)
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
func ensureActionExistsLocally(ctx context.Context, source, target string) error {
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

	dir, err := getActionDirectory(source)
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
func getCustomActionMeta(ctx context.Context, actionDir *dagger.Directory) (*core.CustomActionMeta, error) {
	var meta core.CustomActionMeta

	file, err := findActionMetadataFileName(ctx, actionDir)
	if err != nil {
		return nil, err
	}

	content, err := actionDir.File(file).Contents(ctx)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(content), &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// getActionDirectory returns the directory of the action from given source.
func getActionDirectory(source string) (*dagger.Directory, error) {
	// if path is relative, use the host to resolve the path
	if isLocalAction(source) {
		return config.Client().Host().Directory(source), nil
	}

	// if path is not a relative path, it must be a remote repository in the format "{owner}/{repo}/{path}@{ref}"
	// if {path} is not present in the input string, an empty string is returned for the path component.
	actionRepo, actionPath, actionRef, err := parseRepoRef(source)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repo ref %s: %v", source, err)
	}

	// if path is empty, use the root of the repo as the action directory
	if actionPath == "" {
		actionPath = "."
	}

	// TODO: handle enterprise github instances as well
	// TODO: handle ref type (branch, tag, commit) currently only tags are supported
	return config.Client().Git(path.Join("github.com", actionRepo)).Tag(actionRef).Tree().Directory(actionPath), nil
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
