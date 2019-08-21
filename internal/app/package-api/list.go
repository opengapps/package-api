package packageapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"github.com/opengapps/package-api/internal/pkg/cache"
	"golang.org/x/sync/errgroup"
	"golang.org/x/xerrors"
)

const latestReleaseURLTemplate = "https://raw.githubusercontent.com/opengapps/%s/master/LATEST.json"

// listKey is used for cache
var listKey = cache.NewKey([]byte("list"))

var releaseURLMap = map[gapps.Platform]string{
	gapps.PlatformArm:    fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformArm),
	gapps.PlatformArm64:  fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformArm64),
	gapps.PlatformX86:    fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformX86),
	gapps.PlatformX86_64: fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformX86_64),
}

// LatestRelease describes the latest gapps release
type LatestRelease struct {
	Arch   string         `json:"arch"`
	Date   string         `json:"date"`
	Assets []ReleaseAsset `json:"assets"`
}

// ReleaseAsset describes the gapps release for API and its available variants
type ReleaseAsset struct {
	API      string   `json:"api"`
	Variants []string `json:"variants"`
}

func (a *Application) listHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp ListResponse

		// check the cache first
		cacheValue, ok := a.cache.Get(listKey)
		if ok {
			respond(w, http.StatusOK, cacheValue)
			return
		}

		// prepare query func
		queryFn := func(arch gapps.Platform) func() error {
			return func() error {
				releaseResp, err := http.Get(releaseURLMap[arch])
				if err != nil {
					return xerrors.Errorf("unable to acquire LATEST file for arch '%s': %w", arch, err)
				}
				defer releaseResp.Body.Close()

				if releaseResp.StatusCode != http.StatusOK {
					return xerrors.Errorf("unable to acquire LATEST file for arch '%s': got response '%s'", arch, releaseResp.Status)
				}

				var release LatestRelease
				if err = json.NewDecoder(releaseResp.Body).Decode(&release); err != nil {
					return xerrors.Errorf("unable to decode LATEST file for arch '%s': %w", arch, err)
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

		// async query LATEST files for each arch
		g, _ := errgroup.WithContext(r.Context())
		for _, p := range gapps.PlatformValues() {
			g.Go(queryFn(p))
		}
		if err := g.Wait(); err != nil {
			resp.Error = err.Error()
			respond(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		// save to cache and return the result
		respBody := resp.ToJSON()
		a.cache.Add(listKey, respBody)
		respond(w, http.StatusOK, respBody)
	}
}
