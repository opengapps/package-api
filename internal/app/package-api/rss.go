package packageapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/xerrors"
)

const archAll = "all"

func (a *Application) rssHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get arch from request
		arch, err := parseRSSRequest(r)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		// get all DB recors for the arch
		suffix := ""
		if arch != archAll {
			suffix = arch
		}
		keys, values, err := a.db.GetMultipleBySuffix(suffix)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		// get the sorted records and arch list (required for /all)
		archs, records, err := prepareDBRecords(keys, values)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		// prepare and fill the feed
		feed := &feeds.Feed{
			Title:       fmt.Sprintf(a.cfg.GetString(config.RSSNameKey), arch),
			Description: fmt.Sprintf(a.cfg.GetString(config.RSSDescriptionKey), arch),
			Link: &feeds.Link{
				Type: "application/atom+xml",
				Rel:  "self",
				Href: r.URL.Scheme + a.cfg.GetString(config.APIHostKey) + r.RequestURI,
			},
			Author:    &feeds.Author{Name: a.cfg.GetString(config.RSSAuthorKey)},
			Created:   time.Unix(a.cfg.GetInt64(config.RSSCreationTSKey), 0).UTC(),
			Copyright: a.cfg.GetString(config.RSSCopyrightKey),
			Items:     make([]*feeds.Item, 0, len(records)),
		}

		// fill feed items
		lastUpdated := feed.Created
		for i, record := range records {
			timeCreated := time.Unix(record.Timestamp, 0).UTC()
			link := fmt.Sprintf(a.cfg.GetString(config.RSSLinkKey), archs[i], record.Date)
			feed.Items = append(feed.Items, &feeds.Item{
				Title: fmt.Sprintf(a.cfg.GetString(config.RSSTitleKey), archs[i], record.HumanDate),
				Link: &feeds.Link{
					Href: link,
				},
				Description: fmt.Sprintf(a.cfg.GetString(config.RSSContentKey), record.HumanDate, link),
				Author:      feed.Author,
				Created:     timeCreated,
			})
			if timeCreated.After(lastUpdated) {
				lastUpdated = timeCreated
			}
		}
		feed.Updated = lastUpdated

		atom, err := feed.ToAtom()
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}
		respondXML(w, http.StatusOK, []byte(atom))
	}
}

func prepareDBRecords(keys []string, values [][]byte) ([]string, []db.Record, error) {
	var err error
	archs := make([]string, len(values))
	records := make([]db.Record, len(values))

	for i := 0; i < len(values); i++ {
		record := db.Record{}
		if err = json.Unmarshal(values[i], &record); err != nil {
			return nil, nil, xerrors.Errorf("unable to parse record for key '%s': %w", keys[i], err)
		}
		parts := strings.Split(keys[i], "-")
		if len(parts) != 2 {
			return nil, nil, xerrors.Errorf("unable to parse record for key '%s': bad key", keys[i])
		}
		archs[i] = parts[1]
		records[i] = record
	}
	return archs, records, nil
}

func parseRSSRequest(req *http.Request) (string, error) {
	arch, ok := mux.Vars(req)[queryArgArch]
	if !ok {
		return "", xerrors.Errorf(missingParamErrTemplate, queryArgArch)
	}
	if arch == archAll {
		return archAll, nil
	}

	platform, err := gapps.PlatformString(arch)
	if err != nil {
		return "", xerrors.Errorf("unable to parse '%s' param: '%s' is not a valid architecture", queryArgArch, arch)
	}
	return platform.String(), nil
}
