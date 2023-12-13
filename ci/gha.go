package main

import (
	"context"
	"path/filepath"
	"time"
)

const (
	// ghaPath is the relative path to the home directory of the generated Github Actions modules
	ghaPath = "gha"

	// catalogFile is the name of the catalog file for the generated Github Actions modules
	catalogFile = "catalog.yaml"
)

type Gha struct{}

// Catalog represents the catalog of generated Github Actions modules.
type Catalog struct {
	Global  Global   `yaml:"global,omitempty"`
	Actions []Action `yaml:"actions,omitempty"`
}

// Global represents the global configuration for the generated Github Actions modules.
type Global struct {
	// Global Gale Version to use. If not specified, latest version will be used.
	GaleVersion string `yaml:"galeVersion,omitempty"`

	// Dagger Version to use. If not specified, latest version will be used.
	DaggerVersion string `yaml:"daggerVersion,omitempty"`
}

// Action represents a Github Action configuration to generate.
type Action struct {
	// Name of the action repo. Format:<owner>/<repo>
	Repo string `yaml:"repo"`

	// Version of the action to run. Format:<version>
	Version string `yaml:"version"`

	// Gale Version to use. If not specified, global version will be used.
	GaleVersion string `yaml:"galeVersion,omitempty"`

	// Dagger Version to use. If not specified, global version will be used.
	DaggerVersion string `yaml:"daggerVersion,omitempty"`
}

// Generate generates the Github Actions modules from gha catalog and returns the directory containing the generated modules.
func (g *Gha) Generate(ctx context.Context) (*Directory, error) {
	var (
		gha = dag.Host().Directory(root()).Directory(ghaPath)
		cf  = gha.File(catalogFile)
	)

	var catalog Catalog

	err := unmarshalContentsToYAML(ctx, cf, &catalog)
	if err != nil {
		return nil, err
	}

	for _, action := range catalog.Actions {
		var (
			name   = action.Repo + "@" + action.Version
			gale   = action.GaleVersion
			dagger = action.DaggerVersion
		)

		if gale == "" {
			gale = catalog.Global.GaleVersion
		}

		if dagger == "" {
			dagger = catalog.Global.DaggerVersion
		}

		dir := dag.ActionsGenerator().Generate(name, ActionsGeneratorGenerateOpts{GaleVersion: gale, DaggerVersion: dagger})

		gha = gha.WithDirectory(".", dir)
	}

	return gha, nil
}

// Publish publishes the Github Actions modules from gha catalog to daggerverse.dev
func (g *Gha) Publish(ctx context.Context) error {
	var (
		gha = dag.Host().Directory(root()).Directory(ghaPath)
		cf  = gha.File(catalogFile)
	)

	var catalog Catalog

	err := unmarshalContentsToYAML(ctx, cf, &catalog)
	if err != nil {
		return err
	}

	for _, action := range catalog.Actions {
		daggerVersion := action.DaggerVersion
		if daggerVersion == "" {
			daggerVersion = catalog.Global.DaggerVersion
		}

		_, err = dagger(Opt(daggerVersion)).
			WithMountedDirectory("/src", dag.Host().Directory(root())).
			WithWorkdir("/src").
			WithEnvVariable("CACHE_BUSTER", time.Now().Format(time.RFC3339Nano)).
			WithExec([]string{"mod", "publish", "-f", "-m", filepath.Join(ghaPath, action.Repo)}, ContainerWithExecOpts{ExperimentalPrivilegedNesting: true}).
			Sync(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
