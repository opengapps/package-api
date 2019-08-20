package config_test

import (
	"testing"

	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/stretchr/testify/assert"
)

const (
	testName   = "config"
	testPrefix = "FAKE_ETA_TEST"
)

var testConfigKeys = map[string]interface{}{
	config.DebugKey:      config.DefaultDebugFlag,
	config.CacheLimitKey: config.DefaultCacheLimit,
	config.CacheTTLKey:   config.DefaultCacheTTL,
	config.ServerHostKey: config.DefaultServerHost,
	config.ServerPortKey: config.DefaultServerPort,
}

func TestNew(t *testing.T) {
	cfg := config.New(testName, testPrefix)

	for testKey, testValue := range testConfigKeys {
		assert.Equal(t, cfg.Get(testKey), testValue)
	}
}
