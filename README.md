# GitHub Action Local Executor

The GitHub Action Local Executor, or `gale` for short, is a powerful tool that enables the execution of
[GitHub Actions](https://docs.github.com/en/actions) locally using [Dagger](https://dagger.io).

**Warning**: This project is still in early development and is not ready for use.
**Warning**: This project currently relies on `gh` for authentication. The github cli must be installed and authenticated.

## Features

- Execute GitHub Actions locally
- Test and Debug GitHub Actions locally without pushing to GitHub
- List previous runs of a workflow and access their logs, artifacts, and metadata

## TODO

- [x] Support for `custom` actions using `node`
- [x] Support for `bash` scripts in `run` steps
- [ ] Support for `docker://` actions
- [ ] Support for github expressions, e.g. `${{ github.ref }}`
- [ ] Support for `secrets`
- [ ] Support for triggers and events
- [ ] Support for `composite` actions
- [ ] Support for reusable workflows
- [ ] Add [Default environment variables](https://docs.github.com/en/actions/learn-github-actions/variables#default-environment-variables)

**including list from [aweris/ghx](https://github.com/aweris/ghx) as well

## Installation

```bash
go get github.com/aweris/gale
```

## Usage

### As a CLI

Usage:

```bash
Usage:
  gale [flags]

Flags:
      --export            Export the runner directory after the execution. Exported directory will be placed under .gale directory in the current directory.
  -h, --help              help for gale
      --job string        Name of the job
      --workflow string   Name of the workflow. If workflow doesn't have name, than it must be relative path to the workflow file
```

Run:

```bash
gale --workflow=.github/workflows/clone.yaml --job=clone --export
```

The above command will execute the `clone` job in the workflow `.github/workflows/clone.yaml` and export the runner directory to `.gale/<timestamp>` directory.

### As a library

```go
result, err := gale.New(client).
                WithGithubContext(githubCtx).
                WithJob(".github/workflows/clone.yaml", "clone").
                WithStep(&model.Step{ID: "0", Uses: "actions/checkout@v3", With: map[string]string{"token": jrc.Github.Token}}, true). // override checkout step
                Exec(ctx)
```

The above code will will do same thing as the CLI command above with one difference.  It will override the checkout step with the one provided in the `WithStep` method.