package main

import (
	"fmt"
	"time"
)

// ArtifactCacheEntry represents a single entry in the cache.
//
// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/contracts.d.ts#L9C30-L15
type ArtifactCacheEntry struct {
	CacheKey        string `json:"cacheKey,omitempty"`
	Scope           string `json:"scope,omitempty"`
	CacheVersion    string `json:"cacheVersion,omitempty"`
	CreationTime    string `json:"creationTime,omitempty"`
	ArchiveLocation string `json:"archiveLocation,omitempty"`
}

// ArtifactCacheList represents the response of the artifact cache listing.
//
// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/contracts.d.ts#L17-L20
type ArtifactCacheList struct {
	TotalCount     int                  `json:"totalCount"`
	ArtifactCaches []ArtifactCacheEntry `json:"artifactCaches,omitempty"`
}

// CommitCacheRequest represents the request body of the artifact cache commit request.
//
// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/contracts.d.ts#L22-L24
type CommitCacheRequest struct {
	Size int `json:"size"`
}

// ReserveCacheRequest represents the request body of the artifact cache reserve request.
//
// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/contracts.d.ts#L26-L30
type ReserveCacheRequest struct {
	Key       string `json:"key"`
	Version   string `json:"version,omitempty"`
	CacheSize int    `json:"cacheSize,omitempty"`
}

// ReserveCacheResponse represents the response of the artifact cache reserve request.
//
// source: https://github.com/actions/toolkit/blob/91d3933eb52b351f437151400a88ba7d57442a9b/packages/cache/src/internal/contracts.d.ts#L32-L34
type ReserveCacheResponse struct {
	CacheID uint64 `json:"cacheId"`
}

// CacheEntry represents a single entry in the cache. It is used to store the cache metadata in the database.
type CacheEntry struct {
	ID         uint64 `json:"id"`         // ID is the unique identifier of the cache entry
	Key        string `json:"key"`        // Key is the cache key of the cache entry. It is used to identify the cache entry
	Version    string `json:"version"`    // Version is the version of the cache entry. It is used to identify the cache entry
	Size       int    `json:"size"`       // Size is the size of the cache entry in bytes. It'll be -1 for old actions doesn't support size
	Complete   bool   `json:"complete"`   // Complete indicates whether the cache entry is committed or not
	LastUsedAt int64  `json:"lastUsedAt"` // LastUsedAt is the timestamp of the last time the cache entry is used
	CreatedAt  int64  `json:"createdAt"`  // CreatedAt is the timestamp of the cache entry creation
}

// toArtifactCacheEntry converts the cache entry to an artifact cache entry.
func (c *CacheEntry) toArtifactCacheEntry(hostname string) *ArtifactCacheEntry {
	return &ArtifactCacheEntry{
		CacheKey:        c.Key,
		CacheVersion:    c.Version,
		CreationTime:    time.Unix(c.CreatedAt, 0).UTC().Format(time.RFC3339),
		ArchiveLocation: fmt.Sprintf("http://%s/_apis/artifactcache/artifacts/%d", hostname, c.ID),
	}
}

// updateLastUsedAt updates the last used timestamp of the cache entry.
func (c *CacheEntry) updateLastUsedAt() {
	c.LastUsedAt = time.Now().Unix()
}
