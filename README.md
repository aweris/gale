# GitHub Action Local Executor

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.2-green)

Welcome to project `gale`!

Project `gale` is a Dagger module that allows you to run GitHub Actions locally or anywhere you can run Dagger as if 
they were running on GitHub.

With `gale`, you can get to enjoy a series of perks:

- **Speedy Execution**: Get your workflows running faster compared to the usual GitHub Action Workflow.

- **Cache Support**: Save time on repetitive tasks with enhanced caching features.

- **Programmable Environment**: Customize your execution environment to your liking, making it more adaptable and extensible.

**Heads up:** We are continually working to enhance `gale`. As it is in active development, you might notice changes in 
commands and APIs over time. Your understanding and support are greatly appreciated!

## Before You Begin

### Things You Need

To start using `gale`, make sure your computer has these tools:

1. **Docker**: Dagger, requires [Docker](https://www.docker.com/) running on your host system.

2. **Dagger CLI**: You need to install the Dagger CLI , version >= v0.9.1, to use `gale`. [Install the Dagger CLI](https://docs.dagger.io/quickstart/729236/cli)

Once you have these tools, you are ready to install and use `gale`.

## How to Use

### Setup Dagger Module

To avoid adding `-m github.com/aweris/gale/daggerverse/gale` to every command, you can add run the following command to
set `DAGGER_MODULE` environment variable:

```shell
export DAGGER_MODULE=github.com/aweris/gale/daggerverse/gale
```

### Listing a Workflows

To get a list of workflows, you can use the dagger call workflows list command.

Below is the help output showing the usage and options:

```shell
 Usage:
   dagger call workflows list [flags]

 Aliases:
   list, List

 Flags:
       --branch string          Git branch to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence.
   -h, --help                   help for list
       --repo string            Name of the repository. Format: owner/name.
       --source Directory       Directory containing the repository source. Has precedence over repo.
       --tag string             Git tag to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence.
       --workflows-dir string   Path to the workflow directory. (default ".github/workflows")
```

#### Examples

List all workflows for current repository:

```shell
dagger call workflows list --source "."
```

List workflows for a specific repository and directory:

```shell
dagger call workflows list --repo aweris/gale --branch main --workflows-dir examples/workflows
```

### Run a Workflow

For running workflows, you'll mainly use `dagger [call|download] workflows run [flags] [sub-command]`. Below is
the help output showing the usage and options:

```shell
 Usage:
   dagger call workflows run [flags]
   dagger call workflows run [command]

 Aliases:
   run, Run

 Available Commands:
   config      Configuration for the workflow run.
   directory   Directory returns the directory of the workflow run information.
   sync        Sync evaluates the workflow run and returns the container at executed the workflow.

 Flags:
       --branch string          Git branch to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence.
       --event string           Name of the event that triggered the workflow. (default "push")
       --event-file File        The file with the complete webhook json event payload.
   -h, --help                   help for run
       --job string             Name of the job to run. If empty, all jobs will be run.
       --repo string            Name of the repository. Format: owner/name.
       --runner-debug           Enables debug mode.
       --runner-image string    Docker image to use for the runner. (default "ghcr.io/dagger/gale:latest")
       --source Directory       Directory containing the repository source. Has precedence over repo.
       --tag string             Git tag to checkout. Used with --repo. If tag and branch are both specified, tag takes precedence.
       --token Secret           GitHub token to use for authentication.
       --workflow string        Name of the workflow to run.
```

#### Sub-Commands

- **sync**: Sync evaluates the workflow run and returns the container at executed the workflow.

```shell
dagger call workflow run ... --workflow build sync
```

- **directory**: Exports workflow run data.

The help output for directory option
```shell
Usage:
  dagger download workflows run directory [flags]

Aliases:
  directory, Directory

Flags:
   -h, --help                help for directory
       --include-artifacts   Adds the uploaded artifacts to the exported directory.
       --include-event       Adds the event file to the exported directory.
       --include-repo        Adds the repository source to the exported directory.
       --include-secrets     Adds the mounted secrets to the exported directory.
```

Example command for exporting the workflow run data:
```shell
dagger download workflow run ... --workflow build directory --export-path .gale/exports --include-artifacts
```

##### Examples

Running a workflow for remote repository and downloading exporting the workflow run data and artifacts:

```shell
 dagger download --focus=false workflows run --repo kubernetes/minikube --branch master --workflow build --job build_minikube --token $GITHUB_TOKEN directory --export-path .gale/exports --include-artifacts
```

**Notes for Above Example:**
- `--focus=false` is used to disable focus mode. Required for displaying the execution logs.
- `--token` is optional however it is required for the workflow in this example.

## Feedback and Collaboration

We welcome feedback, suggestions, and collaboration from our users. Your input plays a crucial role in shaping the project and making it even better.

If you encounter any issues, have ideas for improvements, or want to collaborate on this exciting journey, please  feel free to open issues or pull requests on our[ GitHub repository](https://github.com/aweris/gale) or reach out to us on [Discord](https://discord.com/channels/707636530424053791/1117139064274034809)