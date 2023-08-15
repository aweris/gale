package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"go.etcd.io/bbolt"

	"github.com/aweris/gale/internal/log"
)

var (
	// buckets

	dbEntries     = []byte("cache_entries")
	idxVersionKey = []byte("idx_version_key")
)

// BoltStore provides access to BoltDB for the artifact cache store and retrieve cache entries.
type BoltStore struct {
	// conn is the underlying handle to the db.
	conn *bbolt.DB

	// The path to the Bolt database file
	path string
}

// NewBoltStore opens a BoltDB database at the given path and prepares it for use as an artifact cache.
func NewBoltStore(path string) (*BoltStore, error) {
	conn, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	store := &BoltStore{conn: conn, path: path}

	err = store.initialize()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// Close closes the underlying BoltDB database.
func (b *BoltStore) Close() error {
	return b.conn.Close()
}

// initialize creates the buckets needed for the artifact cache.
func (b *BoltStore) initialize() error {
	return b.conn.Update(func(tx *bbolt.Tx) error {

		for _, bucket := range [][]byte{dbEntries, idxVersionKey} {
			_, err := tx.CreateBucketIfNotExists(bucket)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// Exists returns true if the cache entry with the given key and version exists.
func (b *BoltStore) Exists(key, version string) (bool, error) {
	var exists bool

	err := b.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		idx := tx.Bucket(idxVersionKey)
		if idxVersionKey == nil {
			return fmt.Errorf("bucket %s not found", string(idxVersionKey))
		}

		idxKey := fmt.Sprintf("%s:%s", version, key)
		id := idx.Get([]byte(idxKey))
		if id == nil {
			return nil
		}

		val := bucket.Get(id)
		if val == nil {
			return nil
		}

		exists = true

		return nil
	})
	if err != nil {
		return false, err
	}

	return exists, nil
}

// SaveNX saves the given cache entry to the database with next sequence number. If the entry with same
// key and version exists, it will return error.
func (b *BoltStore) SaveNX(entry *CacheEntry) error {
	return b.conn.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		idx := tx.Bucket(idxVersionKey)
		if idxVersionKey == nil {
			return fmt.Errorf("bucket %s not found", string(idxVersionKey))
		}

		// To be able to match key prefixes with same version, we create a key with version and key.
		// this way, we can query the index with the version and key prefix to see if there is a match.
		idxKey := fmt.Sprintf("%s:%s", entry.Version, entry.Key)

		val := idx.Get([]byte(idxKey))
		if val != nil {
			return fmt.Errorf("entry with key %s and version %s already exists", entry.Key, entry.Version)
		}

		id, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		entry.ID = id

		value, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		err = bucket.Put(itob(id), value)
		if err != nil {
			return err
		}

		err = idx.Put([]byte(idxKey), itob(id))
		if err != nil {
			return err
		}

		log.Debugf("Saved a cache entry", "id", id, "entry", string(value))

		return nil
	})
}

// FindByKeyAndVersion returns the cache entry with the given key and version.
func (b *BoltStore) FindByKeyAndVersion(key, version string) (*CacheEntry, error) {
	var entry *CacheEntry

	err := b.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		idx := tx.Bucket(idxVersionKey)
		if idxVersionKey == nil {
			return fmt.Errorf("bucket %s not found", string(idxVersionKey))
		}

		idxKey := fmt.Sprintf("%s:%s", version, key)
		id := idx.Get([]byte(idxKey))
		if id == nil {
			return fmt.Errorf("entry with key %s and version %s not found", key, version)
		}

		val := bucket.Get(id)
		if val == nil {
			return fmt.Errorf("entry with id %d not found", id)
		}

		return json.Unmarshal(val, &entry)
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (b *BoltStore) FindByID(cacheID uint64) (*CacheEntry, error) {
	var entry *CacheEntry

	err := b.conn.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		val := bucket.Get(itob(cacheID))
		if val == nil {
			return fmt.Errorf("entry with id %d not found", cacheID)
		}

		return json.Unmarshal(val, &entry)
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (b *BoltStore) FindByKeyPrefixAnd(key, version string) (*CacheEntry, error) {
	var entry *CacheEntry

	err := b.conn.View(func(tx *bbolt.Tx) error {
		prefix := []byte(fmt.Sprintf("%s:%s", version, key))

		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		idx := tx.Bucket(idxVersionKey)
		if idx == nil {
			return fmt.Errorf("bucket %s not found", string(idxVersionKey))
		}

		c := idx.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			val := bucket.Get(v)
			if val == nil {
				return fmt.Errorf("entry with id %d not found", v)
			}

			var cache *CacheEntry

			if err := json.Unmarshal(val, &cache); err != nil {
				return err
			}

			// skip incomplete entries
			if !cache.Complete {
				continue
			}

			// first match is enough for us
			entry = cache
			break
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Update updates the cache entry with the given key and version.
func (b *BoltStore) Update(entry *CacheEntry) error {
	return b.conn.Update(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(dbEntries)
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", dbEntries)
		}

		if entry.ID == 0 {
			return fmt.Errorf("entry id is required")
		}

		data, err := json.Marshal(entry)
		if err != nil {
			return err
		}

		err = bucket.Put(itob(entry.ID), data)
		if err != nil {
			return err
		}

		return nil
	})
}

// itob returns an 8-byte big endian representation of v.
func itob(u uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, u)
	return buf
}
