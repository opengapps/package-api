package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config keys and default values
const (
	DebugKey            = "debug"
	CacheLimitKey       = "cache_limit"
	CacheTTLKey         = "cache_ttl"
	ServerHostKey       = "server_host"
	ServerPortKey       = "server_port"
	DownloadEndpointKey = "endpoints.download"
	ListEndpointKey     = "endpoints.list"
	MD5EndpointKey      = "endpoints.md5"

	DefaultDebugFlag        = false
	DefaultCacheLimit       = 10000
	DefaultCacheTTL         = "10m"
	DefaultServerHost       = "127.0.0.1"
	DefaultServerPort       = "8080"
	DefaultDLEndpointPath   = "/download"
	DefaultListEndpointPath = "/list"
	DefaultMD5EndpointPath  = "/md5"
)

// New returns new instance of Viper config
func New(name, prefix string) *viper.Viper {
	// try to load config from file first
	cfg := viper.New()
	cfg.SetConfigName(name)
	cfg.AddConfigPath("/etc")
	cfg.AddConfigPath(".")
	if err := cfg.ReadInConfig(); err != nil {
		log.Println("Unable to read config file - using only ENV")
	}

	// add ENV fallback
	cfg.SetEnvPrefix(prefix)
	cfg.AutomaticEnv()

	// set defaults
	cfg.SetDefault(DebugKey, DefaultDebugFlag)
	cfg.SetDefault(CacheLimitKey, DefaultCacheLimit)
	cfg.SetDefault(CacheTTLKey, DefaultCacheTTL)
	cfg.SetDefault(ServerHostKey, DefaultServerHost)
	cfg.SetDefault(ServerPortKey, DefaultServerPort)
	cfg.SetDefault(DownloadEndpointKey, DefaultDLEndpointPath)
	cfg.SetDefault(ListEndpointKey, DefaultListEndpointPath)
	cfg.SetDefault(MD5EndpointKey, DefaultMD5EndpointPath)
	return cfg
}
