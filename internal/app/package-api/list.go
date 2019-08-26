package packageapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	log "github.com/sirupsen/logrus"
)

func (a *Application) listHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := models.ListResponse{
			ArchList: make(map[string]models.ArchRecord, 4),
		}

		// get the releases from the DB
		keys, err := a.db.Keys()
		if err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		for _, p := range gapps.PlatformValues() {
			key := getLatestArchKey(p.String(), keys)
			if key == "" {
				log.Warnf("No releases found for arch '%s'", p)
				continue
			}

			data, err := a.db.Get(key)
			if err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			var record models.ArchRecord
			if err = json.Unmarshal(data, &record); err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			resp.ArchList[p.String()] = record
		}

		if len(resp.ArchList) == 0 {
			respondJSON(w, http.StatusNoContent, nil)
			return
		}

		respondJSON(w, http.StatusOK, resp.ToJSON())
	}
}

func getLatestArchKey(arch string, keys []string) string {
	var result string
	for _, key := range keys {
		parts := strings.Split(key, "-")
		if len(parts) != 2 {
			log.Warnf("Bad DB key '%s'", key)
			continue
		}
		if parts[1] == arch && key > result {
			result = key
		}
	}
	return result
}
