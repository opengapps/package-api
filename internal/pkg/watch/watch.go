package watch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"
	"github.com/opengapps/package-api/internal/pkg/link"
	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/google/go-github/v29/github"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
)

// Watcher periodically queries GitHub for the package changes
type Watcher struct {
	cfg *viper.Viper
	db  *db.DB
	gh  *github.Client
}

// NewWatcher creates new Github release watcher
func NewWatcher(cfg *viper.Viper, storage *db.DB, gh *github.Client) (*Watcher, error) {
	if cfg == nil {
		return nil, errors.New("passed config is nil")
	}
	if storage == nil {
		return nil, errors.New("passed storage is nil")
	}
	if gh == nil {
		return nil, errors.New("passed GitHub client is nil")
	}

	return &Watcher{cfg: cfg, db: storage, gh: gh}, nil
}

// Launch starts the watcher
func (w *Watcher) Launch(ctx context.Context) {
	if err := w.checkRelease(ctx); err != nil {
		log.WithError(err).Error("Unable to check for the latest release")
	}

	period := w.cfg.GetDuration(config.GithubWatchIntervalKey)
	ticker := time.NewTicker(period)

	for {
		select {
		case <-ctx.Done():
			log.Warn("Context canceled, exiting watcher")
			ticker.Stop()
			return
		case <-ticker.C:
			if err := w.checkRelease(ctx); err != nil {
				log.WithError(err).Error("Unable to check for the latest release")
			}
		}
	}
}

func (w *Watcher) checkRelease(ctx context.Context) error {
	var resp models.ListResponse

	g, gCtx := errgroup.WithContext(ctx)
	for _, arch := range gapps.PlatformValues() {
		g.Go(addPackageFn(gCtx, w.gh, &resp, arch))
	}
	if err := g.Wait(); err != nil {
		return err
	}

	// check results and save to DB if necessary
	for arch, record := range resp.ArchList {
		key := fmt.Sprintf(db.KeyTemplate, record.Date, arch)
		_, err := w.db.Get(key)

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
			if err = w.db.Put(key, data); err != nil {
				log.WithError(err).Errorf("Unable to save the data for the arch '%s' and date '%s'", arch, dbRecord.Date)
			}
		default:
			return fmt.Errorf("unable to check DB key: %w", err)
		}
	}

	return nil
}

func addPackageFn(ctx context.Context, gh *github.Client, resp *models.ListResponse, arch gapps.Platform) func() error {
	return func() error {
		release, err := link.GetLatestRelease(ctx, gh, arch)
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
