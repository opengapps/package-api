package app

import (
	"log"

	"github.com/opengapps/package-api/internal/pkg/config"
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

// PrintInfo logs the inforamtion about the current launched binary if Debug mode is enabled)
func PrintInfo(cfg *viper.Viper) {
	if cfg.GetBool(config.DebugKey) {
		log.Print("App info:")
		log.Printf("  Name: %s", Name)
		log.Printf("  Version: %s", Version)
		log.Printf("  Build date: %s", BuildTS)
		log.Printf("  Go version: v%s", GoVersion)

		log.Print("Git info:")
		log.Printf("  Tag: %s", GitBranch)
		log.Printf("  Commit: %s", GitHash)
	}
}
