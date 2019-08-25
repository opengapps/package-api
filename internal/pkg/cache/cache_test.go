package cache_test

import (
	"testing"
	"time"

	"github.com/opengapps/package-api/internal/pkg/cache"
	"github.com/opengapps/package-api/internal/pkg/config"

	"github.com/golang/groupcache/lru"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCacheLimit      = 5
	testEmptyCacheLen   = 0
	testFullCacheLen    = 1
	testCacheTTL        = 20 * time.Millisecond
	testCacheTTLTimeout = 30 * time.Millisecond
	testKeySource       = `test`
)

var (
	testValue      = []byte("")
	testEmptyValue []byte
)

func TestNew_ShouldPass_WithConfig(t *testing.T) {
	cfg := createConfig(t, testCacheLimit, testCacheTTL)

	c, err := cache.New(cfg)

	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestNew_ShouldFail_WithNilConfig(t *testing.T) {
	var cfg *viper.Viper

	c, err := cache.New(cfg)

	assert.Nil(t, c)
	assert.EqualError(t, err, "config is nil")
}

func TestNew_ShouldFail_WithoutCacheLimit(t *testing.T) {
	cfg := createConfig(t, 0, testCacheTTL)

	c, err := cache.New(cfg)

	assert.Nil(t, c)
	assert.EqualError(t, err, "cache limit must be greater than 0")
}

func TestNew_ShouldFail_WithoutCacheTTL(t *testing.T) {
	cfg := createConfig(t, testCacheLimit, 0)

	c, err := cache.New(cfg)

	assert.Nil(t, c)
	assert.EqualError(t, err, "cache TTL must be greater than 0")
}

func TestCacheAdd(t *testing.T) {
	c, key := createEmptyCacheAndKey(t)

	c.Add(key, testValue)

	assert.Equal(t, testFullCacheLen, c.Len())
}

func TestCacheGet_ShouldPass_WithValue(t *testing.T) {
	c, key := createCacheWithValue(t)

	value, ok := c.Get(key)

	assert.True(t, ok)
	assert.Equal(t, testValue, value)
}

func TestCacheGet_ShouldFail_WithoutValue(t *testing.T) {
	c, key := createEmptyCacheAndKey(t)

	value, ok := c.Get(key)

	assert.False(t, ok)
	assert.Equal(t, testEmptyValue, value)
}

func TestCacheGet_ShouldFail_WithValueAfterTTL(t *testing.T) {
	c, key := createCacheWithValue(t)
	time.Sleep(testCacheTTLTimeout)

	value, ok := c.Get(key)

	assert.False(t, ok)
	assert.Equal(t, testEmptyValue, value)
}

func TestCacheRemove(t *testing.T) {
	c, key := createCacheWithValue(t)

	c.Remove(key)

	assert.Equal(t, testEmptyCacheLen, c.Len())
}

func TestCacheClear(t *testing.T) {
	c, _ := createCacheWithValue(t)

	c.Clear()

	assert.Equal(t, testEmptyCacheLen, c.Len())
}

func createConfig(t *testing.T, limit int, ttl time.Duration) *viper.Viper {
	t.Helper()

	cfg := viper.New()
	cfg.Set(config.CacheLimitKey, limit)
	cfg.Set(config.CacheTTLKey, ttl)
	require.NotNil(t, cfg)

	return cfg
}

func createEmptyCacheAndKey(t *testing.T) (*cache.Cache, lru.Key) {
	t.Helper()

	c, err := cache.New(createConfig(t, testCacheLimit, testCacheTTL))
	require.NoError(t, err)
	require.Equal(t, testEmptyCacheLen, c.Len())

	return c, cache.NewKey([]byte(testKeySource))
}

func createCacheWithValue(t *testing.T) (*cache.Cache, lru.Key) {
	t.Helper()

	c, err := cache.New(createConfig(t, testCacheLimit, testCacheTTL))
	require.NoError(t, err)

	key := cache.NewKey([]byte(testKeySource))
	c.Add(key, testValue)
	require.Equal(t, testFullCacheLen, c.Len())

	return c, key
}
