# GitHub Action Local Executor

The GitHub Action Local Executor, or `gale` for short, is a customizable GitHub Action environment that can be run locally.
The goal of this project is to combine the features of the Dagger and GitHub Actions together to create a local
environment that can be used to run GitHub Actions as if they were running on GitHub.

**Please Note:** This project is in active development, commands and APIs subject to change without notice. Frequent updates planned for project improvement.

## Getting Started

### Prerequisites

Before you can use `gale` to execute workflows locally, make sure you have the following tools available on your machine:

1. **[GitHub CLI](https://docs.github.com/en/github-cli/github-cli/quickstart)**:
    - Install and authenticate the GitHub CLI to communicate with the GitHub API.

2. **[Docker](https://www.docker.com/)** and **[Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)**:
    - Dagger, which is utilized by `gale`, requires Docker and Git for running workflows.

Optional:

3. **[Dagger CLI](https://docs.dagger.io/cli/465058/install)**:
    - You may optionally install the Dagger CLI for using `gale` with dagger's TUI

With these prerequisites in place, you'll be ready to set up and run `gale` on your local machine, executing GitHub Actions locally.


### Installation

You can download the latest release of `gale` from the [releases page](https://github.com/aweris/gale/releases). 

or you can install `gale` using:

```bash!
curl -sfLo install.sh https://raw.githubusercontent.com/aweris/gale/main/hack/install.sh; sh ./install.sh
```

The following command will install latest version of `gale` to your current directory. You can also specify following environment variables,

- `GALE_VERSION` to specify the version of `gale` to install, e.g. `GALE_VERSION=v0.0.1` default is `latest`
- `BIN_DIR` to specify the directory to install `gale` to, e.g. `BIN_DIR=/usr/local/bin` default is current directory

```bash!
curl -sfLo install.sh https://raw.githubusercontent.com/aweris/gale/main/hack/install.sh; sudo GALE_VERSION=v0.0.2 BIN_DIR=/usr/local/bin sh ./install.sh
```

**Note:** It's not possible install `gale` using `go install` because `gale` requires the version information to be 
embedded in the binary and `go install` doesn't compile the binary with version information.

### List workflows and jobs

```bash!
dagger run gale list --repo goreleaser/goreleaser --tag v1.19.2
```

This command lists all workflows and jobs under it for the repository (`goreleaser/goreleaser`) and tag (`v1.19.2`). The `--workflows-dir` flag specifies the directory to load workflows from, which, if empty, defaults to `.github/workflows`.

An example execution for the command described above:

![list-workflows](https://github.com/aweris/gale/assets/9319656/ebbf343e-dcac-4942-8570-32f4eced6173)

### Run a job

```bash!
dagger run gale run --workflows-dir ci/workflows ci/workflows/lint.yaml golangci-lint
```

This command runs a specific job (`golangci-lint`) from the workflow defined in `ci/workflows/lint.yaml` file. Since `--repo` 
flag is not specified, repository information of the current directory will be used. 

An example execution for the command described above:

![run-only-gale](https://github.com/aweris/gale/assets/9319656/aa39b487-982a-44e5-b373-17c5aa251d9f)

Since `dagger run` is an optional command, it's omitted in the example above. So we can keep size of the gif small.

### Configuration options

For commands described above all parameters are optional. 

Common parameters for `list` and `run` commands are:

```
  --repo string            owner/repo to load workflows from. If empty, repository information of the current directory will be used.
  --branch string          branch to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
  --tag string             tag to load workflows from. Only one of branch or tag can be used. Precedence is as follows: tag, branch.
  --workflows-dir string   directory to load workflows from. If empty, workflows will be loaded from the default directory.
```

For `run` command, following parameters are also available:

```
  --runner string           runner image to use for running the actions. If empty, the default image will be used.
```
This parameter is useful when you want to use a custom runner image for running the actions. For example, if you want
 to use a custom runner image for running the actions, you can use the following command:

```bash!
dagger run gale run --runner my-custom-runner-image --workflows-dir ci/workflows ci/workflows/lint.yaml golangci-lint
```

Also, running commands with `dagger run` command optional as well but it's preferred execution method to enhance `gale` with `TUI`.

## Feedback and Collaboration

We welcome feedback, suggestions, and collaboration from our users. Your input plays a crucial role in shaping the project and making it even better.

If you encounter any issues, have ideas for improvements, or want to collaborate on this exciting journey, please  feel free to open issues or pull requests on our[ GitHub repository](https://github.com/aweris/gale) or reach out to us on [Discord](https://discord.com/channels/707636530424053791/1117139064274034809)