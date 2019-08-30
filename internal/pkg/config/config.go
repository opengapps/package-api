package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

// Config keys and default values
const (
	APIHostKey             = "api_host"
	ServerHostKey          = "server_host"
	ServerPortKey          = "server_port"
	HTTPTimeoutKey         = "http_timeout"
	DBPathKey              = "db.path"
	DBTimeoutKey           = "db.timeout"
	DownloadEndpointKey    = "endpoint.download"
	ListEndpointKey        = "endpoint.list"
	RSSEndpointKey         = "endpoint.rss"
	GithubTokenKey         = "github.token"
	GithubWatchIntervalKey = "github.watch_interval"

	RSSNameKey        = "rss.name"
	RSSDescriptionKey = "rss.description"
	RSSAuthorKey      = "rss.author"
	RSSCopyrightKey   = "rss.copyright"
	RSSCreationTSKey  = "rss.creation_ts"
	RSSLinkKey        = "rss.link"
	RSSTitleKey       = "rss.title"
	RSSContentKey     = "rss.content"

	DefaultServerHost          = "127.0.0.1"
	DefaultServerPort          = "8080"
	DefaultHTTPTimeout         = "3s"
	DefaultDBPath              = "bolt.db"
	DefaultDBTimeout           = "1s"
	DefaultDLEndpointPath      = "/download"
	DefaultListEndpointPath    = "/list"
	DefaultRSSEndpointPath     = "/rss/{arch}"
	DefaultGithubWatchInterval = "1m"
)

var mandatoryKeys = []string{
	GithubTokenKey,
	RSSNameKey,
	RSSDescriptionKey,
	RSSAuthorKey,
	RSSCopyrightKey,
	RSSCreationTSKey,
	RSSLinkKey,
	RSSTitleKey,
	RSSContentKey,
}

// New returns new instance of Viper config
func New(name, prefix string) (*viper.Viper, error) {
	// try to load config from file first
	cfg := viper.New()
	cfg.SetConfigName(name)
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath(".")
	if err := cfg.ReadInConfig(); err != nil {
		log.WithError(err).Error("Unable to read config file - using only ENV")
	} else {
		// add watch if parsed OK
		cfg.WatchConfig()
	}

	// add ENV fallback
	cfg.SetEnvPrefix(prefix)
	cfg.AutomaticEnv()

	// check mandatory keys
	for _, key := range mandatoryKeys {
		if !cfg.IsSet(key) {
			return nil, xerrors.Errorf("missing mandatory key '%s'", key)
		}
	}

	// set defaults
	cfg.SetDefault(APIHostKey, DefaultServerHost)
	cfg.SetDefault(ServerHostKey, DefaultServerHost)
	cfg.SetDefault(ServerPortKey, DefaultServerPort)
	cfg.SetDefault(HTTPTimeoutKey, DefaultHTTPTimeout)
	cfg.SetDefault(DBPathKey, DefaultDBPath)
	cfg.SetDefault(DBTimeoutKey, DefaultDBTimeout)
	cfg.SetDefault(DownloadEndpointKey, DefaultDLEndpointPath)
	cfg.SetDefault(ListEndpointKey, DefaultListEndpointPath)
	cfg.SetDefault(RSSEndpointKey, DefaultRSSEndpointPath)
	cfg.SetDefault(GithubWatchIntervalKey, DefaultGithubWatchInterval)

	// print contents in debug mode
	log.Debug("Using config:")
	for k, v := range cfg.AllSettings() {
		log.Debugf("  %s: %v", k, v)
	}

	return cfg, nil
}
