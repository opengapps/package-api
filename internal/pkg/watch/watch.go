package watch

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"
	"github.com/opengapps/package-api/internal/pkg/link"
	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/google/go-github/v28/github"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"github.com/spf13/viper"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
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
		return nil, xerrors.New("config is nil")
	}
	if storage == nil {
		return nil, xerrors.New("storage is nil")
	}
	if gh == nil {
		return nil, xerrors.New("GitHub client is nil")
	}

	return &Watcher{cfg: cfg, db: storage, gh: gh}, nil
}

func (w *Watcher) launch(ctx context.Context) {
	period := w.cfg.GetDuration(config.GithubWatchIntervalKey)
	ticker := time.NewTicker(period)

	for {
		select {
		case <-ctx.Done():
			log.Print("Context canceled, exiting watcher")
			ticker.Stop()
			return
		case <-ticker.C:
			if err := w.checkRelease(ctx); err != nil {
				log.Printf("Unable to check for the latest release: %s", err)
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

	// TODO: save to DB
	// for arch, record := range resp.ArchList {
	// 	record.Date
	// }

	return nil
}

func addPackageFn(ctx context.Context, gh *github.Client, resp *models.ListResponse, arch gapps.Platform) func() error {
	return func() error {
		release, err := link.GetLatestRelease(ctx, gh, arch)
		if err != nil {
			return xerrors.Errorf("unable to get latest release: %w", err)
		}

		for _, asset := range release.Assets {
			for _, variant := range asset.Variants {
				pkgAPI, err := gapps.AndroidString(strings.Replace(asset.API, ".", "", -1))
				if err != nil {
					return xerrors.Errorf("unable to parse API '%s' in LATEST file for arch '%s': %w", asset.API, arch, err)
				}

				pkgVariant, err := gapps.VariantString(variant)
				if err != nil {
					return xerrors.Errorf("unable to parse variant '%s' of API '%s' in LATEST file for arch '%s': %w", variant, asset.API, arch, err)
				}

				if err = resp.AddPackage(release.Date, arch, pkgAPI, pkgVariant); err != nil {
					return xerrors.Errorf("unable to add package for variant '%s' of API '%s' in LATEST file for arch '%s': %w", variant, asset.API, arch, err)
				}
			}
		}

		return nil
	}
}
