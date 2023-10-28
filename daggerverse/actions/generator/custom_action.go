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

type action struct {
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

func getCustomAction(source string) (*action, error) {
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
	}
	defer resp.Body.Close()

	// if action.yml does not exist, try action.yaml instead. // FIXME: Find a better way to do this.
	if resp.StatusCode == http.StatusNotFound {
		println(fmt.Sprintf("==> action.yml not found for %s@%s. Trying action.yaml instead...", repo, ref))

		resp, err = http.Get(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/action.yaml", repo, ref))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error downloading action.yaml: %v\n", err)
		}
		// defer already called above, so no need to call again
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error downloading action.yaml: %v", resp.Status)
	}

	contents, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var ca model.CustomActionMeta

	if err := yaml.Unmarshal(contents, &ca); err != nil {
		panic(err)
	}

	return &action{
		Owner:    owner,
		RepoName: name,
		Repo:     repo,
		Ref:      ref,
		Meta:     ca,
	}, nil
}
