package data

import (
	"fmt"
	"github.com/aweris/gale/internal/gctx"
	"path/filepath"

	"dagger.io/dagger"

	"github.com/aweris/gale/internal/config"
	"github.com/aweris/gale/internal/dagger/helpers"
)

// MountPath is the path where the gale data directory is mounted.
const MountPath = "/home/runner/work/_temp/gale"

// Directories names for the gale data directory.
const (
	DirCache     = "cache"
	DirArtifacts = "artifacts"
	DirRuns      = "runs"
	DirActions   = "actions"
)

var _ helpers.WithContainerFuncHook = new(CacheVolume)

// CacheVolume contains the configuration for the cache volume used to persist data for gale.
type CacheVolume struct {
	source *dagger.Directory   // source is identifier of the directory to use as the cache volume's root.
	volume *dagger.CacheVolume // volume is the cache volume.
}

// NewCacheVolume creates a new cache volume.
func NewCacheVolume(repo gctx.RepoContext) *CacheVolume {
	return &CacheVolume{
		source: source(),
		volume: config.Client().CacheVolume(fmt.Sprintf("gale-data-%s-%s", repo.Owner.Login, repo.Name)),
	}
}

func (c *CacheVolume) WithContainerFunc() dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		opts := dagger.ContainerWithMountedCacheOpts{
			Source:  c.source,
			Sharing: dagger.Shared,
		}

		return container.WithMountedCache(MountPath, c.volume, opts)
	}
}

func source() *dagger.Directory {
	return config.Client().
		Directory().
		WithNewDirectory(DirCache).
		WithNewDirectory(DirArtifacts).
		WithNewDirectory(DirRuns).
		WithNewDirectory(DirActions)
}

func (c *CacheVolume) CacheVolume() *dagger.CacheVolume {
	return c.volume
}

func ArtifactsCachePath() string {
	return filepath.Join(MountPath, DirCache)
}

func ArtifactsPath() string {
	return filepath.Join(MountPath, DirArtifacts)
}

func RunsPath() string {
	return filepath.Join(MountPath, DirRuns)
}

func ActionsPath() string {
	return filepath.Join(MountPath, DirActions)
}
