package cache

import (
	"sync"
	"time"

	"github.com/golang/groupcache/lru"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

// Cache is a thread-safe wrapper around groupcache's LRU cache implementation
// It holds positive int64 values with []byte keys
type Cache struct {
	cache *lru.Cache
	ttl   time.Duration
	mtx   sync.RWMutex
}

// New creates new instance of Cache
func New(cfg *viper.Viper) (*Cache, error) {
	if cfg == nil {
		return nil, xerrors.New("config is nil")
	}

	limit := cfg.GetInt(config.CacheLimitKey)
	if limit <= 0 {
		return nil, xerrors.New("cache limit must be greater than 0")
	}
	ttl := cfg.GetDuration(config.CacheTTLKey)
	if ttl <= 0 {
		return nil, xerrors.New("cache TTL must be greater than 0")
	}

	return &Cache{cache: lru.New(limit), ttl: ttl}, nil
}

// Add adds the value to Cache
func (c *Cache) Add(key lru.Key, value []byte) {
	c.mtx.Lock()
	c.cache.Add(key, value)
	c.mtx.Unlock()

	go c.watchTTL(key)
}

// Get acquires value from Cache
func (c *Cache) Get(key lru.Key) ([]byte, bool) {
	c.mtx.RLock()
	defer c.mtx.RUnlock()

	value, ok := c.cache.Get(key)
	if !ok {
		return nil, false
	}

	var v []byte
	v, ok = value.([]byte)
	return v, ok
}

// Remove removes the value from Cache
func (c *Cache) Remove(key lru.Key) {
	c.mtx.Lock()
	c.cache.Remove(key)
	c.mtx.Unlock()
}

// Clear purges the Cache
func (c *Cache) Clear() {
	c.mtx.Lock()
	c.cache.Clear()
	c.mtx.Unlock()
}

// Len returns the current Cache length
func (c *Cache) Len() int {
	c.mtx.RLock()
	defer c.mtx.RUnlock()
	return c.cache.Len()
}

// WatchTTL waits until the TTL is expired and then removes the value of the watched key
func (c *Cache) watchTTL(key lru.Key) {
	time.Sleep(c.ttl)
	c.Remove(key)
}
