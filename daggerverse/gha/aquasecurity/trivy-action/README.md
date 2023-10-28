# Module: Aqua Security Trivy

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.1-green)

Scans container images for vulnerabilities with Trivy

This module is automatically generated using [actions-generator](https://github.com/aweris/gale/tree/main/daggerverse/actions/generator). It is a Dagger-compatible adaptation of the original [aquasecurity/trivy-action](https://github.com/aquasecurity/trivy-action) action.

## How to Use

Run the following command run this action:

```shell
dagger call -m <module-path> run [flags]
```

Replace `<module-path>` with the local path or a git repo reference to the module

## Flags

### Action Inputs

| Name | Required | Description | Default | 
| ------| ------| ------| ------| 
| --with-timeout | false | timeout (default 5m0s) |  |
| --with-scanners | false | comma-separated list of what security issues to detect |  |
| --with-github-pat | false | GitHub Personal Access Token (PAT) for submitting SBOM to GitHub Dependency Snapshot API |  |
| --with-format | false | output format (table, json, template) | table |
| --with-skip-dirs | false | comma separated list of directories where traversal is skipped |  |
| --with-cache-dir | false | specify where the cache is stored |  |
| --with-list-all-pkgs | false | output all packages regardless of vulnerability | false |
| --with-input | false | reference of tar file to scan |  |
| --with-scan-ref | false | Scan reference | . |
| --with-exit-code | false | exit code when vulnerabilities were found |  |
| --with-severity | false | severities of vulnerabilities to be displayed | UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL |
| --with-artifact-type | false | input artifact type (image, fs, repo, archive) for SBOM generation |  |
| --with-ignore-unfixed | false | ignore unfixed vulnerabilities | false |
| --with-vuln-type | false | comma-separated list of vulnerability types (os,library) | os,library |
| --with-template | false | use an existing template for rendering output (@/contrib/gitlab.tpl, @/contrib/junit.tpl, @/contrib/html.tpl) |  |
| --with-output | false | writes results to a file with the specified file name |  |
| --with-hide-progress | false | hide progress output |  |
| --with-trivyignores | false | comma-separated list of relative paths in repository to one or more .trivyignore files |  |
| --with-trivy-config | false | path to trivy.yaml config |  |
| --with-limit-severities-for-sarif | false | limit severities for SARIF format |  |
| --with-scan-type | false | Scan type to use for scanning vulnerability | image |
| --with-image-ref | false | image reference(for backward compatibility) |  |
| --with-skip-files | false | comma separated list of files to be skipped |  |
| --with-ignore-policy | false | filter vulnerabilities with OPA rego language |  |


### Action Runtime Inputs

| Flag | Required | Description | 
| ------| ------| ------| 
| --runner-image | Optional | Image to use for the runner. |
| --runner-debug | Optional | Enables debug mode. |
| --token | Optional | GitHub token is optional for running the action. However, be aware that certain custom actions may require a token and could fail if it's not provided. |
| --source | Conditional | The directory containing the repository source. Either `--source` or `--repo` must be provided; `--source` takes precedence. |
| --repo | Conditional | The name of the repository (owner/name). Either `--source` or `--repo` must be provided; `--source` takes precedence. |
| --tag | Conditional | Tag name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence. |
| --branch | Conditional | Branch name to check out. Only works with `--repo`. Either `--tag` or `--branch` must be provided; `--tag` takes precedence. |
