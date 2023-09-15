# GitHub Action Local Executor

Welcome to project `gale`!

The purpose of this project is to provide an environment for running GitHub Actions as if they were running on GitHub.
The project `gale` harnessed the power of [dagger](https://dagger.io) to create a customizable GitHub Action environment 
that can be run locally or anywhere you can run Dagger.

With `gale`, you can get to enjoy a series of perks:

- **Speedy Execution**: Get your workflows running faster compared to the usual GitHub Action Workflow.

- **Cache Support**: Save time on repetitive tasks with enhanced caching features.

- **Programmable Environment**: Customize your execution environment to your liking, making it more adaptable and extensible.

**Heads up:** We are continually working to enhance `gale`. As it is in active development, you might notice changes in 
commands and APIs over time. Your understanding and support are greatly appreciated!

## Before You Begin

### Things You Need

To start using `gale`, make sure your computer has these tools:

1. **GitHub CLI**: Helps `gale` connect with GitHub. [Learn how to get it here](https://docs.github.com/en/github-cli/github-cli/quickstart).

2. **Docker and Git**: Dagger, which is utilized by `gale`, requires Docker and Git for running workflows. [Find Docker here](https://www.docker.com/) and [Git here](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git).

3. **Dagger CLI (Optional)**: Enhance `gale` with a special TUI from Dagger. [See how to install it here](https://docs.dagger.io/cli/465058/install).

Once you have these tools, you are ready to install and use `gale`.

## How to Install

You can get `gale` in two ways:

1. **Install from GitHub**: You can download the latest release of `gale` from the [releases page](https://github.com/aweris/gale/releases).
2. **Install using script**: Run the following command to install `gale` to your current directory:

```bash!
curl -sfLo install.sh https://raw.githubusercontent.com/aweris/gale/main/hack/install.sh && sh ./install.sh
```

You can customize the installation by specifying the following environment variables:

- `GALE_VERSION` to specify the version of `gale` to install, e.g. `GALE_VERSION=v0.0.1` default is `latest`
- `BIN_DIR` to specify the directory to install `gale` to, e.g. `BIN_DIR=/usr/local/bin` default is current directory

```bash!
curl -sfLo install.sh https://raw.githubusercontent.com/aweris/gale/main/hack/install.sh && sudo GALE_VERSION=v0.0.7 BIN_DIR=/usr/local/bin sh ./install.sh
```

**Note:** You cannot install `gale` with `go install` because this method doesn't include the necessary version information in the compiled binary.

## How to Use

### Viewing Workflows and Jobs

Command: 

```bash!
> gale list --help
List all workflows and jobs under it

Usage:
  gale list [flags]

Flags:
      --branch string          branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
  -h, --help                   help for list
      --repo string            owner/repo to load workflows from. If empty, repository information of the current directory will be used.
      --tag string             tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
      --workflows-dir string   directory to load workflows from. If empty, workflows will be loaded from the default directory.
```

To list all workflows and jobs for the `goreleaser/goreleaser` repository under tag `v1.19.2`, use:

```bash!
dagger run gale list --repo goreleaser/goreleaser --tag v1.19.2
```

This command lists all workflows and jobs under given remote repository and tag. The output is similar to the following:

![list-workflows](https://github.com/aweris/gale/assets/9319656/ebbf343e-dcac-4942-8570-32f4eced6173)

### Running a Workflow

Command: 

```bash!
> gale run --help

Run Github Actions by providing workflow name.

Usage:
  gale run <workflow> [flags]

Flags:
      --branch string           branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
      --debug                   enable debug mode
  -h, --help                    help for run
  -j, --job string              job name to run. if empty, all jobs will be run in the workflow.
      --repo string             owner/repo to load workflows from. If empty, repository information of the current directory will be used.
      --runner string           runner image or path to Dockerfile to use for running the actions. If empty, the default runner image will be used.
      --secret stringToString   secrets to be used in the workflow. Format: --secret name=value (default [])
      --tag string              tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
      --workflows-dir string    directory to load workflows from. If empty, workflows will be loaded from the default directory.
```

To run the `ci/workflows/lint.yaml` workflow for the current directory's repository, use:

```bash!
dagger run gale run --workflows-dir ci/workflows ci/workflows/lint.yaml
```

This commands runs the `lint.yaml` workflow from the `ci/workflows` directory. The output is similar to the following:

![run-only-gale](https://github.com/aweris/gale/assets/9319656/11f580bb-bfc7-4279-9135-48566079437c)

The file path is used as the workflow name since it's not defined like in GitHub 
([Learn more](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#name)). 

Since `dagger run` is an optional command, it's omitted in the example above. So we can keep size of the gif small.

## Feedback and Collaboration

We welcome feedback, suggestions, and collaboration from our users. Your input plays a crucial role in shaping the project and making it even better.

If you encounter any issues, have ideas for improvements, or want to collaborate on this exciting journey, please  feel free to open issues or pull requests on our[ GitHub repository](https://github.com/aweris/gale) or reach out to us on [Discord](https://discord.com/channels/707636530424053791/1117139064274034809)