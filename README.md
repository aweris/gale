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

You can install `gale` by using:

```bash
go install github.com/aweris/gale@latest
```

### List workflows and jobs

```bash!
dagger run gale list --repo aweris/gale --branch main --workflows-dir ci/workflows
```

This command lists the available workflows and jobs in the specified repository (`aweris/gale`) and branch (`main`) located in the `ci/workflows` directory.

<details>
<summary>An example output</summary>

```bash!
➜ dagger run gale list --repo aweris/gale
█ [0.81s] gale list --repo aweris/gale
┃ Workflow: .github/workflows/conventional-label.yaml
┃ Jobs:
┃  - check-for-conventional-release-labels
┃
┃ Workflow: example-gha-run-gale (path: .github/workflows/example-gha-run-gale.yaml)
┃ Jobs:
┃  - run-gale
┃
┃ Workflow: example-golangci-lint (path: .github/workflows/example-golangci-lint.yaml)
┃ Jobs:
┃  - golangci-lint
┃
┃ Workflow: .github/workflows/lint.yaml
┃ Jobs:
┃  - golangci-lint
┃
┃ Workflow: .github/workflows/release-main.yaml
┃ Jobs:
┃  - artifact-service
┃
┃ Workflow: .github/workflows/release-tag.yaml
┃ Jobs:
┃  - gale
┃  - artifact-service
┃
┃ Workflow: .github/workflows/stacked-prs.yaml
┃ Jobs:
┃  - check-for-stacked-prs
█ [0.42s] git://github.com/aweris/gale#main
┃ bfa7fcabc6e9c5a787352dead50579694ded0bfb    refs/heads/main
┻
• Engine: 9224cd98dcf2
• Duration: 808ms
```

</details>

### Run a job

```bash!
dagger run gale run ci/workflows/lint.yaml golangci-lint --repo aweris/gale --branch main --workflows-dir ci/workflows
```

This command runs a specific job (`golangci-lint`) from the workflow defined in `ci/workflows/lint.yaml` in the repository (`aweris/gale`) and branch (`main`). The `--workflows-dir` flag specifies the directory to load workflows from, which, if empty, defaults to `.github/workflows`.

<details>
<summary>An example output</summary>

```bash
➜  gale run ci/workflows/lint.yaml golangci-lint --workflows-dir ci/workflows
3: resolve image config for ghcr.io/aweris/gale/tools/ghx:main
3: > in from ghcr.io/aweris/gale/tools/ghx:main
3: resolve image config for ghcr.io/aweris/gale/tools/ghx:main DONE

5: pull ghcr.io/aweris/gale/tools/ghx:main
5: > in from ghcr.io/aweris/gale/tools/ghx:main
5: resolve ghcr.io/aweris/gale/tools/ghx:main@sha256:c5aa62ac90f1ba7196efa1d4ee262b710fedc7edf92e001b21e102328f318e6f
5: resolve ghcr.io/aweris/gale/tools/ghx:main@sha256:c5aa62ac90f1ba7196efa1d4ee262b710fedc7edf92e001b21e102328f318e6f [0.01s]
5: pull ghcr.io/aweris/gale/tools/ghx:main DONE

5: pull ghcr.io/aweris/gale/tools/ghx:main CACHED
5: > in from ghcr.io/aweris/gale/tools/ghx:main
5: pull ghcr.io/aweris/gale/tools/ghx:main CACHED

10: upload /Users/aweris/development/workspaces/github.com/aweris/gale DONE
10: > in host.directory /Users/aweris/development/workspaces/github.com/aweris/gale
10: upload /Users/aweris/development/workspaces/github.com/aweris/gale DONE

10: upload /Users/aweris/development/workspaces/github.com/aweris/gale
10: > in host.directory /Users/aweris/development/workspaces/github.com/aweris/gale
10: transferring /Users/aweris/development/workspaces/github.com/aweris/gale:
10: transferring /Users/aweris/development/workspaces/github.com/aweris/gale: 29.68MiB [0.69s]
10: upload /Users/aweris/development/workspaces/github.com/aweris/gale DONE

9: copy /Users/aweris/development/workspaces/github.com/aweris/gale CACHED
9: > in host.directory /Users/aweris/development/workspaces/github.com/aweris/gale
9: copy /Users/aweris/development/workspaces/github.com/aweris/gale CACHED

3: resolve image config for ghcr.io/aweris/gale/tools/ghx:main
3: > in from ghcr.io/aweris/gale/tools/ghx:main
3: resolve image config for ghcr.io/aweris/gale/tools/ghx:main DONE

28: resolve image config for ghcr.io/catthehacker/ubuntu:act-22.04
28: > in Runner Base Image > from ghcr.io/catthehacker/ubuntu:act-22.04
28: resolve image config for ghcr.io/catthehacker/ubuntu:act-22.04 DONE

59: mkdir / DONE
59: > in Runner Base Image
59: mkdir / DONE

62: pull ghcr.io/catthehacker/ubuntu:act-22.04
62: > in Runner Base Image > from ghcr.io/catthehacker/ubuntu:act-22.04
62: > in Runner Base Image
62: ...

59: mkdir / CACHED
59: > in Runner Base Image
59: mkdir / CACHED

58: mkfile /job_run.json
58: > in Runner Base Image
58: ...

62: pull ghcr.io/catthehacker/ubuntu:act-22.04 DONE
62: > in Runner Base Image > from ghcr.io/catthehacker/ubuntu:act-22.04
62: > in Runner Base Image
62: resolve ghcr.io/catthehacker/ubuntu:act-22.04@sha256:54d34d7d138215739f4833f0ad08c65d737b5374bf74dec064904efc13b993f1 [0.01s]
62: pull ghcr.io/catthehacker/ubuntu:act-22.04 DONE

58: mkfile /job_run.json DONE
58: > in Runner Base Image
58: mkfile /job_run.json DONE

61: copy /ghx /usr/local/bin/ghx CACHED
61: > in Runner Base Image
61: copy /ghx /usr/local/bin/ghx CACHED

60: merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx) CACHED
60: > in Runner Base Image
60: merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx) CACHED

57: copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e
57: > in Runner Base Image
57: copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e DONE

56: merge (merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx), copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e)
56: > in Runner Base Image
56: merge (merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx), copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e) DONE

56: merge (merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx), copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e)
56: > in Runner Base Image
56: merging
56: merging [0.01s]
56: merge (merge (pull ghcr.io/catthehacker/ubuntu:act-22.04, copy /ghx /usr/local/bin/ghx), copy / /home/runner/_temp/ghx/runs/e17a9fe5-42a4-4d0d-82cb-7787c7002b2e) DONE

54: exec /usr/local/bin/ghx run e17a9fe5-42a4-4d0d-82cb-7787c7002b2e
54: > in Runner Base Image
54: [1.22s] golangci-lint
54: [1.36s] ┏
54: [1.36s] ┃ Set up job
54: [1.36s] ┃ ┏
54: [1.36s] ┃ ┃ Download action repository 'actions/checkout@v3'
54: [1.49s] ┃ ┃ Download action repository 'actions/setup-go@v4'
54: [1.49s] ┃ ┃ Download action repository 'golangci/golangci-lint-action@v3'
54: [1.49s] ┃ ┃ Complete job name: golangci-lint
54: [1.54s] ┃ ┗
54: [1.54s] ┃ Checkout
54: [1.54s] ┃ ┏
54: [1.54s] ┃ ┃ [add-matcher] /home/runner/_temp/ghx/actions/actions/checkout@v3/dist/problem-matcher.json
54: [1.54s] ┃ ┃ Syncing repository: aweris/gale
54: [1.54s] ┃ ┃ Getting Git version info
54: [1.54s] ┃ ┃ ┏
54: [1.54s] ┃ ┃ ┃ Working directory is '/home/runner/work/gale/gale'
54: [1.55s] ┃ ┃ ┃ [command]/usr/bin/git version
54: [1.55s] ┃ ┃ ┃ git version 2.41.0
54: [1.55s] ┃ ┃ ┗
54: [1.55s] ┃ ┃ ...
54: [1.63s] ┃ ┃ Deleting the contents of '/home/runner/work/gale/gale'
54: [1.63s] ┃ ┃ Initializing the repository
54: [1.64s] ┃ ┃ ┏
54: [1.64s] ┃ ┃ ┃ ...
54: [1.64s] ┃ ┃ ┗
54: [1.64s] ┃ ┃ Disabling automatic garbage collection
54: [1.65s] ┃ ┃ ┏
54: [1.65s] ┃ ┃ ┃ [command]/usr/bin/git config --local gc.auto 0
54: [1.65s] ┃ ┃ ┗
54: [1.65s] ┃ ┃ Setting up auth
54: [1.65s] ┃ ┃ ┏
54: [1.65s] ┃ ┃ ┃ ...
54: [1.70s] ┃ ┃ ┗
54: [1.70s] ┃ ┃ Determining the default branch
54: [2.02s] ┃ ┃ ┏
54: [2.02s] ┃ ┃ ┃ Retrieving the default branch name
54: [2.02s] ┃ ┃ ┃ Default branch 'main'
54: [2.02s] ┃ ┃ ┗
54: [2.02s] ┃ ┃ Fetching the repository
54: [2.67s] ┃ ┃ ┏
54: [2.67s] ┃ ┃ ┃ ...
54: [2.92s] ┃ ┃ ┗
54: [2.92s] ┃ ┃ Determining the checkout info
54: [2.92s] ┃ ┃ ┏
54: [2.92s] ┃ ┃ ┗
54: [2.92s] ┃ ┃ Checking out the ref
54: [2.93s] ┃ ┃ ┏
54: [2.93s] ┃ ┃ ┃ [command]/usr/bin/git checkout --progress --force -B main refs/remotes/origin/main
54: [2.93s] ┃ ┃ ┃ Switched to a new branch 'main'
54: [2.93s] ┃ ┃ ┃ branch 'main' set up to track 'origin/main'.
54: [2.94s] ┃ ┃ ┗
54: [2.94s] ┃ ┃ [command]/usr/bin/git log -1 --format='%H'
54: [2.95s] ┃ ┃ 'bfa7fcabc6e9c5a787352dead50579694ded0bfb'
54: [3.02s] ┃ ┗
54: [3.02s] ┃ Setup Go
54: [3.02s] ┃ ┏
54: [3.63s] ┃ ┃ Setup go version spec 1.20
54: [3.65s] ┃ ┃ Attempting to download 1.20...
54: [4.93s] ┃ ┃ matching 1.20...
54: [4.93s] ┃ ┃ Not found in manifest.  Falling back to download directly from Go
54: [8.31s] ┃ ┃ Install from dist
54: [8.31s] ┃ ┃ Acquiring go1.20.6 from https://storage.googleapis.com/golang/go1.20.6.linux-arm64.tar.gz
54: [9.67s] ┃ ┃ Extracting Go...
54: [9.67s] ┃ ┃ [command]/usr/bin/tar xz --warning=no-unknown-keyword --overwrite -C /home/runner/_temp/73e38a8c-8a5d-4ba8-9dd5-88df757ca10e -f /home/runner/_temp/0d047768-bc01-413f-b5ed-2034257809c9
54: [9.67s] ┃ ┃ Successfully extracted go to /home/runner/_temp/73e38a8c-8a5d-4ba8-9dd5-88df757ca10e
54: [13.2s] ┃ ┃ Adding to the cache ...
54: [13.2s] ┃ ┃ Successfully cached go to /opt/hostedtoolcache/go/1.20.6/arm64
54: [13.2s] ┃ ┃ Added go to the path
54: [13.2s] ┃ ┃ Successfully set up Go version 1.20
54: [13.2s] ┃ ┃ [warn] The runner was not able to contact the cache service. Caching will be skipped
54: [13.2s] ┃ ┃ [add-matcher] /home/runner/_temp/ghx/actions/actions/setup-go@v4/matchers.json
54: [13.2s] ┃ ┃ go version go1.20.6 linux/arm64
54: [13.2s] ┃ ┃
54: [13.2s] ┃ ┃ go env
54: [13.2s] ┃ ┃ ┏
54: [13.2s] ┃ ┃ ┃ ...
54: [13.3s] ┃ ┃ ┗
54: [13.3s] ┃ ┗
54: [13.3s] ┃ golangci-lint
54: [13.3s] ┃ ┏
54: [13.3s] ┃ ┃ prepare environment
54: [13.3s] ┃ ┃ ┏
54: [13.3s] ┃ ┃ ┃ ...
54: [14.2s] ┃ ┃ ┗
54: [22.4s] ┃ ┃ run golangci-lint
54: [22.4s] ┃ ┃ ┏
54: [22.4s] ┃ ┃ ┃ Running [/root/golangci-lint-1.53.3-linux-arm64/golangci-lint run --out-format=github-actions] in [] ...
54: [22.4s] ┃ ┃ ┃ golangci-lint found no issues
54: [22.4s] ┃ ┃ ┃ Ran golangci-lint in 8212ms
54: [22.5s] ┃ ┃ ┗
54: [22.5s] ┃ ┗
54: [22.7s] ┃ Complete job
54: [22.7s] ┃ ┏
54: [22.7s] ┃ ┃ Complete job name: golangci-lint conclusion=success
54: [22.7s] ┃ ┗
54: [22.7s] ┗
54: exec /usr/local/bin/ghx run e17a9fe5-42a4-4d0d-82cb-7787c7002b2e DONE
```

</details>

### Configuration options

For commands described above all parameters are optional. Available command flags for `list` and `run` commands are:
```
  --branch string          branch to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.
  --commit string          commit to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.
  --repo string            owner/repo to load workflows from. If empty, repository information of the current directory will be used.
  --tag string             tag to load workflows from. Only one of branch, tag or commit can be used. Precedence is as follows: commit, tag, branch.
  --workflows-dir string   directory to load workflows from. If empty, workflows will be loaded from the default directory.
```

Also, running commands with `dagger run` command optional as well but it's preferred execution method to enhance `gale` with `TUI`.

## Feedback and Collaboration

We welcome feedback, suggestions, and collaboration from our users. Your input plays a crucial role in shaping the project and making it even better.

If you encounter any issues, have ideas for improvements, or want to collaborate on this exciting journey, please  feel free to open issues or pull requests on our[ GitHub repository](https://github.com/aweris/gale) or reach out to us on [Discord](https://discord.com/channels/707636530424053791/1117139064274034809)