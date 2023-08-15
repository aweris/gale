package main

import (
	"errors"
	"io"
	"path/filepath"
	"strconv"
	"time"

	galefs "github.com/aweris/gale/internal/fs"
	"github.com/aweris/gale/internal/log"
)

var (
	// errors

	ErrCacheEntryComplete    = errors.New("cache already committed")
	ErrCacheEntryNotComplete = errors.New("cache not committed")
)

type Service interface {
	// Exist if an artifact cache entry exists for the given key and version.
	Exist(key, version string) (bool, error)

	// Reserve reserves a cache entry for the given key. The cache entry will be created if it does not exist,
	// otherwise it will return error.
	Reserve(key, version string, size int) (uint64, error)

	// Upload uploads the given data to the cache entry with the given id. The data will be appended to the cache
	// entry. The offset indicates the offset of the data in the cache entry.
	Upload(cacheID int, offset int, reader io.Reader) error

	// Commit commits the cache entry with the given id. The cache entry will be marked as complete and will not
	// accept any more data.
	Commit(cacheID int) error

	// Find finds the cache entry with the given key and version. If the cache entry is found, it will return true
	// and the cache entry. Otherwise, it will return false and nil.
	//
	// In list of keys, the first key is always exact match and the rest are partial matches. The partial matches
	// are used to find the latest cache entry with given restore-keys.
	Find(keys []string, version string) (bool, *ArtifactCacheEntry, error)

	// GetFilePath returns the path to the cache entry with the given id. If the cache entry is not found or not
	// complete, it will return error.
	GetFilePath(cacheID int) (string, error)
}

var (
	_ Service   = new(LocalService)
	_ io.Closer = new(LocalService)
)

type LocalService struct {
	path     string     // path to the artifact cache directory
	db       *BoltStore // db to store cache entries
	hostname string     // hostname is the external hostname of the artifact cache service. It is used to construct the artifact cache URL
	port     string     // port is the external port of the artifact cache service. It is used to construct the artifact cache URL
}

// TODO: add cache entry expiration / cleanup. Currently, the cache entries are never deleted. Only when dagger cache volume is deleted, the cache entries are deleted.

// NewLocalService creates a new local artifact service.
func NewLocalService(root, hostname, port string) (*LocalService, error) {
	db, err := NewBoltStore(filepath.Join(root, "metadata"))
	if err != nil {
		return nil, err
	}

	return &LocalService{
		db:       db,
		path:     root,
		hostname: hostname,
		port:     port,
	}, nil
}

// Close closes the local artifact service.
func (s *LocalService) Close() error {
	return s.db.Close()
}

// Exist if an artifact cache entry exists for the given key and version.
func (s *LocalService) Exist(key, version string) (bool, error) {
	ok, err := s.db.Exists(key, version)
	if err != nil {
		log.Errorf("Failed to check cache existence", "error", err, "key", key, "version", version)
		return false, err
	}

	log.Debugf("Check cache existence", "key", key, "version", version, "exists", ok)

	return ok, nil
}

func (s *LocalService) Reserve(key, version string, size int) (uint64, error) {
	if size == 0 {
		size = -1
	}

	now := time.Now().Unix()

	entry := &CacheEntry{
		ID:         0,
		Key:        key,
		Version:    version,
		Size:       size,
		Complete:   false,
		LastUsedAt: now,
		CreatedAt:  now,
	}

	if err := s.db.SaveNX(entry); err != nil {
		log.Errorf("Failed to save cache entry", "error", err, "key", key, "version", version)
		return 0, err
	}

	log.Debugf("Reserved cache entry", "key", key, "version", version, "id", entry.ID)

	return entry.ID, nil
}

func (s *LocalService) Upload(cacheID, offset int, reader io.Reader) error {
	entry, err := s.db.FindByID(uint64(cacheID))
	if err != nil {
		log.Debugf("Failed to find cache entry", "error", err, "id", cacheID)
		return err
	}

	if entry.Complete {
		log.Debugf("Cache entry already committed", "id", cacheID)
		return ErrCacheEntryComplete
	}

	writer, err := galefs.NewMultipartFileWriter(s.getCacheDir(cacheID))
	if err != nil {
		log.Errorf("Failed to create multipart file writer", "error", err, "id", cacheID)
		return err
	}

	err = writer.Write("archive", offset, reader)
	if err != nil {
		log.Errorf("Failed to write data to multipart file writer", "error", err, "id", cacheID)
		return err
	}

	log.Debugf("Uploaded data to cache entry", "id", cacheID, "offset", offset)

	return nil
}

func (s *LocalService) Commit(cacheID int) error {
	entry, err := s.db.FindByID(uint64(cacheID))
	if err != nil {
		log.Debugf("Failed to find cache entry", "error", err, "id", cacheID)
		return err
	}

	if entry.Complete {
		log.Debugf("Cache entry already committed", "id", cacheID)
		return ErrCacheEntryComplete
	}

	entry.Complete = true
	entry.updateLastUsedAt()

	writer, err := galefs.NewMultipartFileWriter(s.getCacheDir(cacheID))
	if err != nil {
		log.Errorf("Failed to create multipart file writer", "error", err, "id", cacheID)
		return err
	}

	if err := writer.Merge(); err != nil {
		log.Errorf("Failed to merge multipart file writer", "error", err, "id", cacheID)
		return err
	}

	err = s.db.Update(entry)
	if err != nil {
		log.Errorf("Failed to update cache entry", "error", err, "id", cacheID)
		return err
	}

	log.Debugf("Committed cache entry", "id", cacheID)

	return nil
}

func (s *LocalService) Find(keys []string, version string) (bool, *ArtifactCacheEntry, error) {
	log.Debugf("Finding cache entry", "keys", keys, "version", version)

	// exact match
	key := keys[0]

	ok, err := s.db.Exists(key, version)
	if err != nil {
		log.Errorf("Failed to check cache existence", "error", err, "key", key, "version", version)
		return false, nil, err
	}

	if ok {
		entry, err := s.db.FindByKeyAndVersion(key, version)
		if err != nil {
			log.Errorf("Failed to find cache entry", "error", err, "key", key, "version", version)
			return false, nil, err
		}

		log.Debugf("Found cache entry", "key", key, "version", version, "id", entry.ID)

		return true, entry.toArtifactCacheEntry(s.hostname, s.port), nil
	}

	// partial match

	for _, key := range keys[1:] {
		entry, err := s.db.FindByKeyPrefixAnd(key, version)
		if err != nil {
			log.Errorf("Failed to find cache entry with key prefix", "error", err, "key", key, "version", version)

			return false, nil, err
		}

		if entry != nil {
			log.Debugf("Found cache entry with key prefix", "key", key, "version", version, "entryID", entry.ID, "entryKey", entry.Key)

			return true, entry.toArtifactCacheEntry(s.hostname, s.port), nil
		}
	}

	log.Debugf("Cache entry not found", "keys", keys, "version", version)

	return false, nil, nil
}

func (s *LocalService) GetFilePath(cacheID int) (string, error) {
	entry, err := s.db.FindByID(uint64(cacheID))
	if err != nil {
		log.Debugf("Failed to find cache entry", "error", err, "id", cacheID)
		return "", err
	}

	if !entry.Complete {
		log.Debugf("Cache entry not committed", "id", cacheID)
		return "", ErrCacheEntryNotComplete
	}

	entry.updateLastUsedAt()

	if err := s.db.Update(entry); err != nil {
		log.Errorf("Failed to update cache entry", "error", err, "id", cacheID)
		return "", err
	}

	path := s.getCacheFilePath(cacheID)

	log.Debugf("Found cache entry", "id", cacheID, "path", path)

	return path, nil
}

func (s *LocalService) getCacheDir(cacheID int) string {
	return filepath.Join(s.path, strconv.Itoa(cacheID))
}

func (s *LocalService) getCacheFilePath(cacheID int) string {
	return filepath.Join(s.path, strconv.Itoa(cacheID), "archive")
}
