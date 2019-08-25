package packageapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/opengapps/package-api/internal/pkg/cache"
	"github.com/opengapps/package-api/internal/pkg/link"
	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/google/go-github/v28/github"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

// listKey is used for cache
var listKey = cache.NewKey([]byte("list"))

func (a *Application) listHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp models.ListResponse

		// check the cache first
		cacheValue, ok := a.cache.Get(listKey)
		if ok {
			respondJSON(w, http.StatusOK, cacheValue)
			return
		}

		// async query LATEST files for each arch
		g, ctx := errgroup.WithContext(r.Context())
		for _, p := range gapps.PlatformValues() {
			g.Go(prepareQuery(ctx, a.gh, &resp, p))
		}
		if err := g.Wait(); err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		// save to cache and return the result
		respBody := resp.ToJSON()
		a.cache.Add(listKey, respBody)
		respondJSON(w, http.StatusOK, respBody)
	}
}

func prepareQuery(ctx context.Context, gh *github.Client, resp *models.ListResponse, arch gapps.Platform) func() error {
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
