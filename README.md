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

## Installation

```bash
go get github.com/aweris/gale
```

## Usage

```bash
Usage:
  gale [command]

Available Commands:
  build       Build a Runner image
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  run         Run a workflow

Flags:
  -h, --help   help for gale

Use "gale [command] --help" for more information about a command.
```
