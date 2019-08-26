package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Config keys and default values
const (
	APIHostKey             = "api_host"
	ServerHostKey          = "server_host"
	ServerPortKey          = "server_port"
	DBPathKey              = "db.path"
	DBTimeoutKey           = "db.timeout"
	DownloadEndpointKey    = "endpoint.download"
	ListEndpointKey        = "endpoint.list"
	RSSEndpointKey         = "endpoint.rss"
	GithubTokenKey         = "github.token"
	GithubWatchIntervalKey = "github.watch_interval"

	DefaultServerHost          = "127.0.0.1"
	DefaultServerPort          = "8080"
	DefaultDBPath              = "bolt.db"
	DefaultDBTimeout           = "1s"
	DefaultDLEndpointPath      = "/download"
	DefaultListEndpointPath    = "/list"
	DefaultRSSEndpointPath     = "/rss"
	DefaultGithubWatchInterval = "1m"
)

// New returns new instance of Viper config
func New(name, prefix string) *viper.Viper {
	// try to load config from file first
	cfg := viper.New()
	cfg.SetConfigName(name)
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath(".")
	if err := cfg.ReadInConfig(); err != nil {
		log.WithError(err).Error("Unable to read config file - using only ENV")
	}

	// add ENV fallback
	cfg.SetEnvPrefix(prefix)
	cfg.AutomaticEnv()

	// set defaults
	cfg.SetDefault(APIHostKey, DefaultServerHost)
	cfg.SetDefault(ServerHostKey, DefaultServerHost)
	cfg.SetDefault(ServerPortKey, DefaultServerPort)
	cfg.SetDefault(DBPathKey, DefaultDBPath)
	cfg.SetDefault(DBTimeoutKey, DefaultDBTimeout)
	cfg.SetDefault(DownloadEndpointKey, DefaultDLEndpointPath)
	cfg.SetDefault(ListEndpointKey, DefaultListEndpointPath)
	cfg.SetDefault(RSSEndpointKey, DefaultRSSEndpointPath)
	cfg.SetDefault(GithubWatchIntervalKey, DefaultGithubWatchInterval)
	return cfg
}
