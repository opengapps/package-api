package config_test

import (
	"testing"

	"github.com/opengapps/package-api/internal/pkg/config"

	"github.com/stretchr/testify/assert"
)

const (
	testName   = "config"
	testPrefix = "PACKAGE_API_TEST"
)

var testConfigKeys = map[string]interface{}{
	config.APIHostKey:             config.DefaultServerHost,
	config.ServerHostKey:          config.DefaultServerHost,
	config.ServerPortKey:          config.DefaultServerPort,
	config.DBPathKey:              config.DefaultDBPath,
	config.DBTimeoutKey:           config.DefaultDBTimeout,
	config.DownloadEndpointKey:    config.DefaultDLEndpointPath,
	config.ListEndpointKey:        config.DefaultListEndpointPath,
	config.GithubWatchIntervalKey: config.DefaultGithubWatchInterval,
}

func TestNew(t *testing.T) {
	cfg := config.New(testName, testPrefix)

	for testKey, testValue := range testConfigKeys {
		assert.Equal(t, cfg.Get(testKey), testValue)
	}
}
