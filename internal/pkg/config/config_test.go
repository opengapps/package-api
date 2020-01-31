package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/opengapps/package-api/internal/pkg/config"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testName        = "config"
	testPrefix      = "PACKAGE_API"
	testValueString = "test"
	testValueInt    = "123"
)

var testConfigKeys = map[string]interface{}{
	config.APIHostKey:             config.DefaultServerHost,
	config.ServerHostKey:          config.DefaultServerHost,
	config.ServerPortKey:          config.DefaultServerPort,
	config.DBPathKey:              config.DefaultDBPath,
	config.DBTimeoutKey:           config.DefaultDBTimeout,
	config.DownloadEndpointKey:    config.DefaultDLEndpointPath,
	config.ListEndpointKey:        config.DefaultListEndpointPath,
	config.RSSEndpointKey:         config.DefaultRSSEndpointPath,
	config.GithubWatchIntervalKey: config.DefaultGithubWatchInterval,
	config.RSSHistoryLengthKey:    config.DefaultRSSHistoryLength,
}

var testConfigEnvs = map[string]string{
	config.AuthKey:           testValueString,
	config.GithubTokenKey:    testValueString,
	config.RSSNameKey:        testValueString,
	config.RSSDescriptionKey: testValueString,
	config.RSSAuthorKey:      testValueString,
	config.RSSCopyrightKey:   testValueString,
	config.RSSCreationTSKey:  testValueInt,
	config.RSSLinkKey:        testValueString,
	config.RSSTitleKey:       testValueString,
	config.RSSContentKey:     testValueString,
}

func TestNew(t *testing.T) {
	setupConfigEnv()

	cfg, err := config.New(testName, testPrefix)

	require.NoError(t, err)
	for testKey, testValue := range testConfigKeys {
		assert.Equal(t, cfg.Get(testKey), testValue)
	}
}

func setupConfigEnv() {
	log.SetLevel(log.FatalLevel) // ignore config logging for tests
	for k, v := range testConfigEnvs {
		os.Setenv(testPrefix+"_"+strings.ToUpper(k), v)
	}
}
