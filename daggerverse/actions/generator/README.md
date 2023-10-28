# Actions Generator

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.1-green)

Generate Dagger modules from Github Actions([Custom Actions](https://docs.github.com/en/actions/creating-actions/about-custom-actions))

## Prerequisites

- Module requires Dagger CLI version `v0.9.1` or higher.

## Before you start

You can set `DAGGER_MODULE` to environment variable to avoid using `-m github.com/aweris/gale/daggerverse/actions/generator` in every command.

```shell
export DAGGER_MODULE=github.com/aweris/gale/daggerverse/actions/generator
```

## Commands

### Generate Dagger Module

```shell
dagger download generate --action <action> --export-path <export-path> [optional flags]
```

#### Flags:

- `--action`: Action repository and version to generate Dagger module from. It is in the format of `<action-repo>@<version>`.
- `--export-path`: Path to export generated Dagger module. It will be exported under `<export-path>/<action-repo>` directory.
- `--runtime-version` (optional): The actions/runtime to execute the action on dagger. 
- `--dagger-version` (optional): The dagger version to use for the generated module. Defaults to latest version.