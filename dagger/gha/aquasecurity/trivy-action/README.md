# Module: Aqua Security Trivy

![dagger-min-version](https://img.shields.io/badge/dagger%20version-v0.9.5-green)

Scans container images for vulnerabilities with Trivy

This module is automatically generated using [actions-generator](https://github.com/aweris/gale/tree/main/daggerverse/actions/generator). It is a Dagger-compatible adaptation of the original [aquasecurity/trivy-action](https://github.com/aquasecurity/trivy-action) action.

## How to Use

Run the following command run this action:

```shell
dagger call -m github.com/aweris/gale/gha/aquasecurity/trivy-action run [flags]
```

## Flags

### Action Inputs

| Name | Required | Description | Default | 
| ------| ------| ------| ------| 
| artifact-type | false | input artifact type (image, fs, repo, archive) for SBOM generation |  |
| cache-dir | false | specify where the cache is stored |  |
| exit-code | false | exit code when vulnerabilities were found |  |
| format | false | output format (table, json, template) | table |
| github-pat | false | GitHub Personal Access Token (PAT) for submitting SBOM to GitHub Dependency Snapshot API |  |
| hide-progress | false | hide progress output |  |
| ignore-policy | false | filter vulnerabilities with OPA rego language |  |
| ignore-unfixed | false | ignore unfixed vulnerabilities | false |
| image-ref | false | image reference(for backward compatibility) |  |
| input | false | reference of tar file to scan |  |
| limit-severities-for-sarif | false | limit severities for SARIF format |  |
| list-all-pkgs | false | output all packages regardless of vulnerability | false |
| output | false | writes results to a file with the specified file name |  |
| scan-ref | false | Scan reference | . |
| scan-type | false | Scan type to use for scanning vulnerability | image |
| scanners | false | comma-separated list of what security issues to detect |  |
| severity | false | severities of vulnerabilities to be displayed | UNKNOWN,LOW,MEDIUM,HIGH,CRITICAL |
| skip-dirs | false | comma separated list of directories where traversal is skipped |  |
| skip-files | false | comma separated list of files to be skipped |  |
| template | false | use an existing template for rendering output (@/contrib/gitlab.tpl, @/contrib/junit.tpl, @/contrib/html.tpl) |  |
| tf-vars | false | path to terraform tfvars file |  |
| timeout | false | timeout (default 5m0s) |  |
| trivy-config | false | path to trivy.yaml config |  |
| trivyignores | false | comma-separated list of relative paths in repository to one or more .trivyignore files |  |
| vuln-type | false | comma-separated list of vulnerability types (os,library) | os,library |


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
