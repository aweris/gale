# Artifact Server

Github Action compatible artifact server to store and retrieve artifacts. This service meant to be used as a
Dagger service binding when running Gale in a non Github Actions environment.

## Usage

### Configuration

The following configuration options are available:

| Flag             | Environment Variable | Description                     | Default      |
|------------------|----------------------|---------------------------------|--------------|
| `--port`         | `PORT`               | Port to listen on               | `8080`       |
| `--artifact-dir` | `ARTIFACT_DIR`       | Directory to store artifacts in | `/artifacts` |

