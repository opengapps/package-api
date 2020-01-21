package db

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/opengapps/package-api/internal/pkg/models"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

const (
	// KeyTemplate describes the format of the keys inside of a bucket
	KeyTemplate = "%s-%s"

	openMode = 0755
)

// Package vars
var (
	ErrNotFound = errors.New("key not found")
	ErrNilValue = errors.New("value is nil")

	bucketName = []byte("global")
)

// Record is used to store models.ArchRecord in DB
type Record struct {
	models.ArchRecord

	Disabled  bool  `json:"disabled,omitempty"`
	Timestamp int64 `json:"ts"`
}

// DB describes local BoltDB database
type DB struct {
	b       *bbolt.DB
	timeout time.Duration
}

// New creates new instance of DB
func New(path string, timeout time.Duration) (*DB, error) {
	// open connection to the DB
	log.WithField("path", path).WithField("timeout", timeout).Debug("Creating DB connection")
	opts := bbolt.DefaultOptions
	if timeout > 0 {
		opts.Timeout = timeout
	}
	b, err := bbolt.Open(path, openMode, opts)
	if err != nil {
		return nil, fmt.Errorf("unable to open DB: %w", err)
	}

	// create global bucket if it doesn't exist yet
	log.WithField("bucket", string(bucketName)).Debug("Setting the default bucket")
	err = b.Update(func(tx *bbolt.Tx) error {
		_, bErr := tx.CreateBucketIfNotExists(bucketName)
		return bErr
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create global bucket: %w", err)
	}

	// return the DB
	db := &DB{b: b, timeout: timeout}
	log.Debug("DB initiated")
	return db, nil
}

// Close closes the DB
func (db *DB) Close(delete bool) error {
	log.Debug("Closing the DB")
	path := db.b.Path()
	done := make(chan error)
	go func() {
		done <- db.b.Close()
		log.Debug("DB closed OK")
		close(done)
	}()
	timer := time.NewTimer(db.timeout)
	if delete {
		defer os.Remove(path)
	}

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("unable to close DB: %w", err)
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("unable to close DB: %w", bbolt.ErrTimeout)
	}
}

// Keys returns a list of available keys in the global bucket, sorted alphabetically
func (db *DB) Keys() ([]string, error) {
	var keys []string
	log.Debug("Getting the list of DB current keys")
	err := db.b.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.ForEach(func(k, v []byte) error {
			if v != nil {
				keys = append(keys, string(k))
			}
			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get the list of keys from DB: %w", err)
	}
	log.Debug("Got the keys")
	sort.Strings(keys)
	return keys, nil
}

// Get acquires value from DB by provided key
func (db *DB) Get(key string) ([]byte, error) {
	var value []byte
	log.WithField("key", key).Debug("Getting value from DB")
	err := db.b.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		k, v := b.Cursor().Seek([]byte(key))
		if k == nil || string(k) != key {
			return ErrNotFound
		} else if v == nil {
			return ErrNilValue
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get value for key '%s' from DB: %w", key, err)
	}
	log.WithField("key", key).Debug("Got the value")
	return value, nil
}

// GetMultipleBySuffix returns keys and non-empty values, for which the key contains the suffix
// It returns all keys and values from the bucket if the suffix is empty
func (db *DB) GetMultipleBySuffix(suffix string) ([]string, [][]byte, error) {
	var (
		keys   []string
		values [][]byte
	)
	log.WithField("suffix", suffix).Debug("Getting values from DB by suffix")
	err := db.b.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.ForEach(func(k, v []byte) error {
			if (suffix == "" || bytes.HasSuffix(k, []byte(suffix))) && v != nil {
				keys = append(keys, string(k))
				values = append(values, v)
			}
			return nil
		})
	})
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get values for suffix '%s' from DB: %w", suffix, err)
	}
	log.WithField("suffix", suffix).Debug("Got the values")
	return keys, values, nil
}

// Put sets/updates the value in DB by provided bucket and key
func (db *DB) Put(key string, val []byte) error {
	log.WithField("key", key).Debug("Saving the value to DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Put([]byte(key), val)
	})
	if err != nil {
		return fmt.Errorf("unable to put value for key '%s' to DB: %w", key, err)
	}
	log.WithField("key", key).Debug("Saved successfully")
	return nil
}

// Delete removes the value from DB by provided bucket and key
func (db *DB) Delete(key string) error {
	log.WithField("key", key).Debug("Deleting from DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)
		if b == nil {
			return bbolt.ErrBucketNotFound
		}
		return b.Delete([]byte(key))
	})
	if err != nil {
		return fmt.Errorf("unable to delete value for key '%s' from DB: %w", key, err)
	}
	log.WithField("key", key).Debug("Deleted successfully")
	return nil
}

// Purge removes the bucket from DB
func (db *DB) Purge() error {
	log.Debug("Purging the DB")
	err := db.b.Update(func(tx *bbolt.Tx) error {
		return tx.DeleteBucket(bucketName)
	})
	if err != nil {
		return fmt.Errorf("unable to purge global bucket from DB: %w", err)
	}
	return nil
}
