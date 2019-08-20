package cache

import (
	"crypto/sha256"

	"github.com/golang/groupcache/lru"
)

// NewKey creates new sha256-encoded byte key for the Cache
func NewKey(body []byte) lru.Key {
	return sha256.Sum256(body)
}
