package packageapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opengapps/package-api/internal/pkg/db"
	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	log "github.com/sirupsen/logrus"
)

func (a *application) listHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := models.ListResponse{
			ArchList: make(map[string]models.ArchRecord, 4),
		}

		// get the release keys from the DB
		keys, err := a.storage.Keys()
		if err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		for _, p := range gapps.PlatformValues() {
			// get the record from the DB and add it to the response
			record, err := a.getLatestRecord(p.String(), keys, []string{})
			if err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			if record == nil {
				log.Warnf("No releases found for arch '%s'", p)
				continue
			}

			resp.ArchList[p.String()] = record.ArchRecord
		}

		if len(resp.ArchList) == 0 {
			respondJSON(w, http.StatusNoContent, nil)
			return
		}
		respondJSON(w, http.StatusOK, resp.ToJSON())
	}
}

func (a *application) getLatestRecord(arch string, keys, disabledKeys []string) (*db.Record, error) {
	key := getLatestArchKey(arch, keys, disabledKeys)
	if key == "" {
		return nil, nil
	}

	data, err := a.storage.Get(key)
	if err != nil {
		return nil, err
	}

	var record db.Record
	if err = json.Unmarshal(data, &record); err != nil {
		return nil, err
	}

	if record.Disabled {
		return a.getLatestRecord(arch, keys, append(disabledKeys, key))
	}

	return &record, nil
}

func getLatestArchKey(arch string, keys, disabledKeys []string) string {
	var result string
	for _, key := range keys {
		parts := strings.Split(key, "-") // date-arch
		if len(parts) != 2 {
			log.Warnf("Bad DB key '%s'", key)
			continue
		}
		// latest date by models.DateOnlyFormat is always bigger
		if parts[1] == arch && !isStringInSlice(key, disabledKeys) && key > result {
			result = key
		}
	}
	return result
}

func isStringInSlice(s string, ss []string) bool {
	for i := range ss {
		if ss[i] == s {
			return true
		}
	}
	return false
}
