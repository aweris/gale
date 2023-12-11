# GitHub Action Local Executor

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.4-green)

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

2. **Dagger CLI**: You need to install the Dagger CLI , version >= v0.9.4, to use `gale`. [Install the Dagger CLI](https://docs.dagger.io/quickstart/729236/cli)

Once you have these tools, you are ready to install and use `gale`.

## How to Use

### Setup Dagger Module

To avoid adding `-m github.com/aweris/gale` to every command, you can add run the following command to
set `DAGGER_MODULE` environment variable:

```shell
export DAGGER_MODULE=github.com/aweris/gale
```

### Global Flags for `gale`

The following table lists the global flags available for use with the `gale` module:

| Flag              | Condition   | Description                                                                              |
|-------------------|-------------|------------------------------------------------------------------------------------------|
| `--repo`          | Conditional | Specify the repository in 'owner/name' format. Used if `--source` is not provided.       |
| `--source`        | Conditional | Path to the repository source directory. Takes precedence over `--repo`.                 |
| `--tag`           | Conditional | Tag name to check out. Takes precedence over `--branch`. Required when `--repo` is used. |
| `--branch`        | Conditional | Branch name to check out. Used if `--tag` is not specified and `--repo` is used.         |
| `--workflows-dir` | Optional    | Path to the workflows' directory. Defaults to `.github/workflows`.                       |

### Listing a Workflows

To get a list of workflows, you can use the dagger call workflows list command.

Below is the help output showing the usage and options:

```shell
List returns a list of workflows and their jobs with the given options.

Usage:
  dagger call list [flags]

Flags:
  -h, --help                   help for list
```

#### Examples

List all workflows for current repository:

```shell
dagger -m github.com/aweris/gale call --source "." list 
```

List workflows for a specific repository and directory:

```shell
dagger -m github.com/aweris/gale call --repo aweris/gale --branch main --workflows-dir examples/workflows list
```

### Run a Workflow

For running workflows, you'll mainly use `dagger [call|download] run [flags] [sub-command]`. Below is
the help output showing the usage and options:

```shell
Run runs the workflow with the given options.

Usage:
  dagger call run [flags]
  dagger call run [command]

Available Commands:
   data        Returns the directory containing the workflow run data.
   log         Returns all job run logs as a single file.
   sync        Returns the container for the given job id. If there is on one job in the workflow run, then job id is not required.

 Flags:
       --container Container   Container to use for the runner(default: ghcr.io/catthehacker/ubuntu:act-latest).
       --docker-host string    Sets DOCKER_HOST to use for the native docker support. (default "unix:///var/run/docker.sock")
       --event string          Name of the event that triggered the workflow. e.g. push (default "push")
       --event-file File       File with the complete webhook event payload.
   -h, --help                  help for run
       --job string            Name of the job to run. If empty, all jobs will be run.
       --runner-debug          Enables debug mode.
       --token Secret          GitHub token to use for authentication.
       --use-dind              Enables docker-in-dagger support to be able to run docker commands isolated from the host. Enabling DinD may lead to longer execution times.
       --use-native-docker     Enables native Docker support, allowing direct execution of Docker commands in the workflow. (default true)
       --workflow string       Name of the workflow to run.
       --workflow-file File    External workflow file to run.
```

##### Examples

Running a workflow for remote repository and downloading exporting the workflow run data and artifacts:

```shell
 dagger -m github.com/aweris/gale export --repo kubernetes/minikube --branch master run --workflow build --job build_minikube --token $GITHUB_TOKEN data --output .gale/exports
```

**Notes for Above Example:**
- `--token` is optional however it is required for the workflow in this example.

## Feedback and Collaboration

We welcome feedback, suggestions, and collaboration from our users. Your input plays a crucial role in shaping the project and making it even better.

If you encounter any issues, have ideas for improvements, or want to collaborate on this exciting journey, please  feel free to open issues or pull requests on our[ GitHub repository](https://github.com/aweris/gale) or reach out to us on [Discord](https://discord.com/channels/707636530424053791/1117139064274034809)