// package-api enables user to interact with Open GApps packages.
// This is mainly used as the opengapps.org backend.
package main

import (
	"log"

	"github.com/opengapps/package-api/internal/app"
	packageapi "github.com/opengapps/package-api/internal/app/package-api"
	"github.com/opengapps/package-api/internal/pkg/cache"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/spf13/pflag"
)

var configName string

func init() {
	pflag.StringVarP(&configName, "config", "c", app.Name, "Config name")
	pflag.Parse()
	log.Printf("Using config name '%s'", configName)
}

func main() {
	log.Println("Initiating the service")

	// init config from ENV
	cfg := config.New(configName, app.Name)

	// init cache
	cache, err := cache.New(cfg)
	if err != nil {
		log.Fatalf("Unable to init cache: %s", err)
	}

	// create and run the server
	a, err := packageapi.New(cfg, cache)
	if err != nil {
		log.Fatalf("Unable to init application: %s", err)
	}
	defer func() {
		if cErr := a.Close(); cErr != nil {
			log.Printf("Error on server shutdown: %s", cErr)
		}
	}()

	log.Println("Starting the server")
	if err = a.Run(); err != nil {
		log.Fatalf("Unable to start the server: %s", err)
	}
}
