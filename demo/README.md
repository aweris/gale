# Demo

Welcome to the interactive demo showcasing the versatility of `gale` in various scenarios.

## Demos: 

Explore `gale` through the following scenarios:

| Demo            | Flag to run         | Description                                                                                     |
|-----------------|---------------------|-------------------------------------------------------------------------------------------------|
| List            | `--list`            | List all workflows and jobs under it for current repositories `main` branch                     |
| Run             | `--run`             | Run golangci-lint job from ci/workflows/lint workflow for aweris/gale repository default branch |
| Lint GoReleaser | `--lint-goreleaser` | Run golangci job from golangci-lint workflow for goreleaser/goreleaser repository tag v1.19.2   |
| Test Dagger     | `--test-dagger`     | Run sdk-go job from test workflow for dagger/dagger repository tag v0.8.1                       |
| Test Cache      | `--test-cache`      | Use actions/cache in the workflow                                                               |

## Getting Started

Execute the demo using the following command from the root of the gale repository:

```bash
go run ./demo/ <demo-flag> --auto --auto-timeout 1s
```

Replace <demo-flag> with the desired flag corresponding to the demo you want to explore.

## How It Works

Upon execution, the chosen demo will perform the following steps:

- Downloading Binaries: The demo will automatically download `dagger` and `gale` binaries with predefined versions into the `./bin` directory.
- Running Demo: The demo will run the chosen demo using the `dagger` and `gale` binaries.
- Cleaning Up: The demo will clean up the `./bin` directory.


This demo uses [saschagrunert/demo](https://github.com/saschagrunert/demo) library to create the interactive demo. Please refer to the [documentation](https://github.com/saschagrunert/demo#usage) for more information on demo running options.