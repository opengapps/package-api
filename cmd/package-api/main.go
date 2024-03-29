// package-api enables user to interact with Open GApps packages.
// This is mainly used as the opengapps.org backend.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/opengapps/package-api/internal/app"
	packageapi "github.com/opengapps/package-api/internal/app/package-api"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"
	"github.com/opengapps/package-api/pkg/github"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var configName string

func init() {
	// get flags, init logger
	pflag.StringVarP(&configName, "config", "c", app.Name, "Config file name")
	level := pflag.String("log-level", "INFO", "Logrus log level (DEBUG, WARN, etc.)")
	pflag.Parse()

	logLevel, err := log.ParseLevel(*level)
	if err != nil {
		log.Errorf("Unknown log level: %s", *level)
		pflag.PrintDefaults()
		os.Exit(1)
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(logLevel)
	log.Debug("Enabling debug logging")

	if configName == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	log.Info("Initiating the service")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// init config from ENV
	cfg, err := config.New(configName, app.Name)
	if err != nil {
		log.WithError(err).Fatal("Unable to init config")
	}

	// init storage
	log.Debug("Initiating DB")
	storage, err := db.New(cfg.GetString(config.DBPathKey), cfg.GetDuration(config.DBTimeoutKey))
	if err != nil {
		log.WithError(err).Fatal("Unable to init storage")
	}

	// init Github client
	log.Debug("Creating Github client")
	githubClient, err := github.NewClient(
		ctx,
		github.WithConfig(cfg),
		github.WithStorage(storage),
	)
	if err != nil {
		log.WithError(err).Fatal("Unable to init Github client")
	}
	go githubClient.Watch(ctx)

	// create the server
	log.Debug("Creating the app server")
	a, err := packageapi.New(
		packageapi.WithConfig(cfg),
		packageapi.WithStorage(storage),
	)
	if err != nil {
		log.WithError(err).Fatal("Unable to init application")
	}

	// init graceful stop chan
	log.Debug("Initiating system signal watcher")
	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		sig := <-gracefulStop
		log.Warnf("Caught sig %+v, stopping the app", sig)
		cancel()
		if err = storage.Close(false); err != nil {
			log.WithError(err).Error("Unable to close DB")
		}
		if err = a.Close(); err != nil {
			log.WithError(err).Error("Error on server shutdown")
		}
		log.Info("Shutting down")
		os.Exit(0)
	}()

	log.Info("Starting the server")
	if err = a.Run(); err != nil {
		log.WithError(err).Fatal("Unable to start the server")
	}
}
