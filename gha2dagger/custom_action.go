package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/common/model"
)

type CustomAction struct {
	// Repo is the repository name in the form "owner/repo".
	Repo string

	// RepoName is the repository name in the form "repo".
	RepoName string

	// Owner is the repository owner.
	Owner string

	// Ref is the git ref of the custom action.
	Ref string

	// Meta contains action.yml contents for the custom action.
	Meta model.CustomActionMeta
}

func NewCustomAction(source string) (*CustomAction, error) {
	regex := regexp.MustCompile(`^([^/]+)/([^/@]+)(?:/([^@]+))?@(.+)$`)
	matches := regex.FindStringSubmatch(source)

	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid input format: %q", source)
	}

	var (
		owner = matches[1]
		name  = matches[2]
		repo  = strings.Join([]string{matches[1], matches[2]}, "/")
		ref   = matches[4]
	)

	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.yml", repo, ref))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error downloading action.yml: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var ca model.CustomActionMeta

	if err := yaml.Unmarshal(contents, &ca); err != nil {
		panic(err)
	}

	return &CustomAction{
		Owner:    owner,
		RepoName: name,
		Repo:     repo,
		Ref:      ref,
		Meta:     ca,
	}, nil
}
