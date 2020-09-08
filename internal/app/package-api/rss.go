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
)

const archAll = "all"

func (a *application) rssHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get arch from request
		arch, err := parseRSSRequest(r)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, errToBytes(err))
			return
		}

		// get all DB recors for the arch
		suffix := ""
		if arch != archAll {
			suffix = arch
		}
		keys, values, err := a.storage.GetMultipleBySuffix(suffix)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, errToBytes(err))
			return
		}

		// get the sorted records and arch list (required for /all)
		archs, records, err := prepareDBRecords(keys, values)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, errToBytes(err))
			return
		}

		// prepare and fill the feed
		scheme := "http://"
		if r.TLS != nil || a.cfg.GetBool(config.HTTPSRedirectKey) {
			scheme = "https://"
		}
		feed := &feeds.Feed{
			Title:       fmt.Sprintf(a.cfg.GetString(config.RSSNameKey), arch),
			Id:          scheme + a.cfg.GetString(config.APIHostKey) + r.RequestURI,
			Description: fmt.Sprintf(a.cfg.GetString(config.RSSDescriptionKey), arch),
			Link: &feeds.Link{
				Type: "application/atom+xml",
				Rel:  "self",
				Href: scheme + a.cfg.GetString(config.APIHostKey) + r.RequestURI,
			},
			Author:    &feeds.Author{Name: a.cfg.GetString(config.RSSAuthorKey)},
			Created:   time.Unix(a.cfg.GetInt64(config.RSSCreationTSKey), 0).UTC(),
			Copyright: a.cfg.GetString(config.RSSCopyrightKey),
			Items:     make([]*feeds.Item, 0, len(records)),
		}

		// fill feed items
		lastUpdated := feed.Created
		firstDay := time.Now().AddDate(0, -a.cfg.GetInt(config.RSSHistoryLengthKey), 0).UTC()
		for i := len(records) - 1; i >= 0; i-- {
			timeCreated := time.Unix(records[i].Timestamp, 0).UTC()
			if timeCreated.Before(firstDay) {
				break // we show only last 'RSSHistoryLength' months of data
			}
			link := fmt.Sprintf(a.cfg.GetString(config.RSSLinkKey), archs[i], records[i].Date)
			feed.Items = append(feed.Items, &feeds.Item{
				Title: fmt.Sprintf(a.cfg.GetString(config.RSSTitleKey), archs[i], records[i].HumanDate),
				Link: &feeds.Link{
					Href: link,
				},
				Description: fmt.Sprintf(a.cfg.GetString(config.RSSContentKey), records[i].HumanDate, link),
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
			respond(w, "", http.StatusInternalServerError, errToBytes(err))
			return
		}
		respondXML(w, http.StatusOK, []byte(atom))
	}
}

func prepareDBRecords(keys []string, values [][]byte) ([]string, []db.Record, error) {
	var err error
	archs := make([]string, 0, len(values))
	records := make([]db.Record, 0, len(values))

	for i := 0; i < len(values); i++ {
		record := db.Record{}
		if err = json.Unmarshal(values[i], &record); err != nil {
			return nil, nil, fmt.Errorf("unable to parse record for key '%s': %w", keys[i], err)
		}
		parts := strings.Split(keys[i], "-")
		if len(parts) != 2 {
			return nil, nil, fmt.Errorf("unable to parse record for key '%s': bad key", keys[i])
		}
		// we ignore disabled records
		if record.Disabled {
			continue
		}
		archs = append(archs, parts[1])
		records = append(records, record)
	}
	return archs, records, nil
}

func parseRSSRequest(req *http.Request) (string, error) {
	arch, ok := mux.Vars(req)[queryArgArch]
	if !ok {
		return "", fmt.Errorf(missingParamErrTemplate, queryArgArch)
	}
	if arch == archAll {
		return archAll, nil
	}

	platform, err := gapps.PlatformString(arch)
	if err != nil {
		return "", fmt.Errorf("unable to parse '%s' param: '%s' is not a valid architecture", queryArgArch, arch)
	}
	return platform.String(), nil
}
