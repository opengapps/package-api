// package-api enables user to interact with Open GApps packages.
// This is mainly used as the opengapps.org backend.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/opengapps/package-api/internal/app"
	packageapi "github.com/opengapps/package-api/internal/app/package-api"
	"github.com/opengapps/package-api/internal/pkg/cache"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"

	"github.com/google/go-github/v28/github"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
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

	// init Github client
	log.Print("Creating Github client")
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GetString(config.GithubTokenKey)},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	gh := github.NewClient(tc)

	// init cache
	cache, err := cache.New(cfg)
	if err != nil {
		log.Fatalf("Unable to init cache: %s", err)
	}

	// init storage
	storage, err := db.New(cfg.GetString(config.DBPathKey), cfg.GetDuration(config.DBTimeoutKey))
	if err != nil {
		log.Fatalf("Unable to init storage: %s", err)
	}

	// create and run the server
	a, err := packageapi.New(cfg, cache, storage, gh)
	if err != nil {
		log.Fatalf("Unable to init application: %s", err)
	}

	// init graceful stop chan
	log.Print("Initiating system signal watcher")
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		log.Printf("Caught sig %+v, stopping the app", sig)
		if err = storage.Close(false); err != nil {
			log.Printf("Unable to close DB: %s", err)
		}
		if cErr := a.Close(); cErr != nil {
			log.Printf("Error on server shutdown: %s", cErr)
		}
		log.Println("Shutting down")
		os.Exit(0)
	}()

	log.Println("Starting the server")
	if err = a.Run(); err != nil {
		log.Fatalf("Unable to start the server: %s", err)
	}
}
