package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/opengapps/package-api/pkg/gapps"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"
	"github.com/opengapps/package-api/internal/pkg/models"
)

type client struct {
	cfg     *viper.Viper
	client  *github.Client
	storage Storage

	once sync.Once
}

// NewClient creates new Github client
func NewClient(ctx context.Context, opts ...Option) (*client, error) {
	c := &client{}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("unable to create client: %w", err)
		}
	}
	if c.cfg == nil {
		return nil, errors.New("config is nil")
	}
	if c.storage == nil {
		return nil, errors.New("storage is nil")
	}
	if c.client == nil {
		return nil, errors.New("client for Github is nil")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.cfg.GetString(config.GithubTokenKey)},
	)
	c.client = github.NewClient(oauth2.NewClient(ctx, ts))

	return c, nil
}

func (c *client) Watch(ctx context.Context) {
	c.once.Do(func() { c.watch(ctx) })
}

// watch starts the release watcher
func (c *client) watch(ctx context.Context) {
	if err := c.checkRelease(ctx); err != nil {
		log.WithError(err).Error("Unable to check for the latest release")
	}

	period := c.cfg.GetDuration(config.GithubWatchIntervalKey)
	ticker := time.NewTicker(period)

	for {
		select {
		case <-ctx.Done():
			log.Warn("Context canceled, exiting watcher")
			ticker.Stop()
			return
		case <-ticker.C:
			if err := c.checkRelease(ctx); err != nil {
				log.WithError(err).Error("Unable to check for the latest release")
			}
		}
	}
}

func (c *client) checkRelease(ctx context.Context) error {
	var resp models.ListResponse

	g, gCtx := errgroup.WithContext(ctx)
	for _, arch := range gapps.PlatformValues() {
		g.Go(c.addPackageFn(gCtx, &resp, arch))
	}
	if err := g.Wait(); err != nil {
		return err
	}

	// check results and save to DB if necessary
	for arch, record := range resp.ArchList {
		key := fmt.Sprintf(db.KeyTemplate, record.Date, arch)
		_, err := c.storage.Get(key)

		switch {
		case err == nil:
			// data is already there, continue
			continue
		case errors.Is(err, db.ErrNilValue), errors.Is(err, db.ErrNotFound):
			// save the new data
			var data []byte
			dbRecord := db.Record{ArchRecord: record, Timestamp: time.Now().Unix()}
			if data, err = json.Marshal(dbRecord); err != nil {
				log.WithError(err).Errorf("Unable to marshal the data for the arch '%s' and date '%s'", arch, dbRecord.Date)
				continue
			}
			if err = c.storage.Put(key, data); err != nil {
				log.WithError(err).Errorf("Unable to save the data for the arch '%s' and date '%s'", arch, dbRecord.Date)
			}
		default:
			return fmt.Errorf("unable to check DB key: %w", err)
		}
	}

	return nil
}

func (c *client) addPackageFn(ctx context.Context, resp *models.ListResponse, arch gapps.Platform) func() error {
	return func() error {
		release, err := c.GetLatestRelease(ctx, arch)
		if err != nil {
			return err
		}

		for _, asset := range release.Assets {
			for _, variant := range asset.Variants {
				pkgAPI, err := gapps.AndroidString(strings.Replace(asset.API, ".", "", -1))
				if err != nil {
					return fmt.Errorf("unable to parse API '%s' in LATEST file for arch '%s': %w", asset.API, arch, err)
				}

				pkgVariant, err := gapps.VariantString(variant)
				if err != nil {
					return fmt.Errorf("unable to parse variant '%s' of API '%s' in LATEST file for arch '%s': %w", variant, asset.API, arch, err)
				}

				if err = resp.AddPackage(release.Date, arch, pkgAPI, pkgVariant); err != nil {
					return fmt.Errorf("unable to add package for variant '%s' of API '%s' in LATEST file for arch '%s': %w", variant, asset.API, arch, err)
				}
			}
		}

		return nil
	}
}
