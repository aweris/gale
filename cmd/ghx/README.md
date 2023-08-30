# GitHub Actions Executor

The GitHub Actions Executor is a helper command for running [GitHub Actions](https://docs.github.com/en/actions) from
inside of [Dagger](https://dagger.io) containers.

## Usage

### Global Configuration

The following configuration options are available:

| Flag     | Environment Variable | Description            | Default                  |
|----------|----------------------|------------------------|--------------------------|
| `--home` | `GHX_HOME`           | home directory for ghx | `/home/runner/_temp/ghx` |