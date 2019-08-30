package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// App vars, overridden by ldflags
var (
	Name      = "package-api"
	Version   = "devel"
	BuildTS   = "_"
	GoVersion = "_"
	GitHash   = "_"
	GitBranch = "_"
)

// PrintInfo logs the information about the current launched binary on DEBUG lvl
func PrintInfo(cfg *viper.Viper) {
	log.Debug("App info:")
	log.Debugf("  Name: %s", Name)
	log.Debugf("  Version: %s", Version)
	log.Debugf("  Build date: %s", BuildTS)
	log.Debugf("  Go version: v%s", GoVersion)

	log.Debug("Git info:")
	log.Debugf("  Tag: %s", GitBranch)
	log.Debugf("  Commit: %s", GitHash)
}
