// Cache service mimics the GitHub Actions cache service. It is used to cache and restore files and directories between
// workflow runs.
// Implementation is based on the actions/toolkit cache client:
// https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/cacheHttpClient.ts
package main
