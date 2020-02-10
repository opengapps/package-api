package packageapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"

	"github.com/opengapps/package-api/internal/pkg/db"
)

const (
	actionEnable    = "enable"
	actionDisable   = "disable"
	gappsDateFormat = "20060102"
)

type pkgRequest struct {
	Action   string `json:"action"`
	Platform string `json:"platform,omitempty"`
	Date     string `json:"date"`
}

// Validate checks if the package request fields are valid
func (r *pkgRequest) Validate() error {
	if r == nil {
		return errors.New("request is empty")
	}

	if r.Action != actionEnable && r.Action != actionDisable {
		return errors.New("bad Action value")
	}

	if r.Platform != "" {
		if _, err := gapps.PlatformString(r.Platform); err != nil {
			return err
		}
	}

	if _, err := time.Parse(gappsDateFormat, r.Date); err != nil {
		return errors.New("bad Date format")
	}

	return nil
}

type pkgResponse struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// ToJSON forms JSON body from a struct, ignoring Marshal error
func (r *pkgResponse) ToJSON() []byte {
	body, _ := json.Marshal(r)
	return body
}

func (a *Application) pkgHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// unmarshal and validate request
		req := &pkgRequest{}
		resp := &pkgResponse{}

		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(req); err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		if err := req.Validate(); err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		// get the release keys from the DB
		allKeys, err := a.db.Keys()
		if err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		var keys []string
		for i := range allKeys {
			prefix := req.Date
			if req.Platform != "" {
				prefix += "-" + req.Platform
			}
			if strings.HasPrefix(allKeys[i], prefix) {
				keys = append(keys, allKeys[i])
			}
		}
		if len(keys) == 0 {
			resp.Error = "package with such date was not found"
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		for _, key := range keys {
			data, err := a.db.Get(key)
			if err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			var record db.Record
			if err = json.Unmarshal(data, &record); err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			switch req.Action {
			case actionEnable:
				if record.Disabled {
					record.Disabled = false
				}
			case actionDisable:
				if !record.Disabled {
					record.Disabled = true
				}
			}

			if data, err = json.Marshal(record); err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			if err = a.db.Put(key, data); err != nil {
				resp.Error = err.Error()
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}
		}

		resp.Status = "OK"
		respondJSON(w, http.StatusOK, resp.ToJSON())
	}
}
