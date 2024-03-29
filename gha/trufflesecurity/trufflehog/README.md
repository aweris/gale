# Module: TruffleHog OSS

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.5-green)

Scan Github Actions with TruffleHog.

This module is automatically generated using [actions-generator](https://github.com/aweris/gale/tree/main/daggerverse/actions/generator). It is a Dagger-compatible adaptation of the original [trufflesecurity/trufflehog](https://github.com/trufflesecurity/trufflehog) action.

## How to Use

Run the following command run this action:

```shell
dagger call -m github.com/aweris/gale/gha/trufflesecurity/trufflehog run [flags]
```

## Flags

### Action Inputs

| Name | Required | Description | Default | 
| ------| ------| ------| ------| 
| base | false | Start scanning from here (usually main branch). |  |
| extra_args | false | Extra args to be passed to the trufflehog cli. |  |
| head | false | Scan commits until here (usually dev branch). |  |
| path | false | Repository path | ./ |


### Action Runtime Inputs

| Flag | Required | Description | 
| ------| ------| ------| 
| --branch | Conditional | Branch name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence. |
| --container | Optional | Container to use for the runner. |
| --repo | Conditional | The name of the repository (owner/name). Either `--source` or `--repo` must be provided; `--source` takes precedence. |
| --runner-debug | Optional | Enables debug mode. |
| --source | Conditional | The directory containing the repository source. Either `--source` or `--repo` must be provided; `--source` takes precedence. |
| --tag | Conditional | Tag name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence. |
| --token | Optional | GitHub token is optional for running the action. However, be aware that certain custom actions may require a token and could fail if it's not provided. |
