# Artifact Cache Server

Github Action compatible cache server to store and retrieve caches. This service meant to be used as a 
Dagger service binding when running Gale in a non Github Actions environment.

## Usage

### Configuration

The following configuration options are available:

| Flag                  | Environment Variable | Description                                | Default         |
|-----------------------|----------------------|--------------------------------------------|-----------------|
| `--port`              | `PORT`               | Port to listen on                          | `8080`          |
| `--cache-dir`         | `CACHE_DIR`          | Directory to store caches in               | `/caches`       |
| `--external-hostname` | `EXTERNAL_HOSTNAME`  | External hostname to use for download URLs | `artifactcache` |