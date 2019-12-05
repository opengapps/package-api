package packageapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"github.com/opengapps/package-api/internal/pkg/db"
	log "github.com/sirupsen/logrus"
)

const (
	actionEnable    = "enable"
	actionDisable   = "disable"
	gappsDateFormat = "20060102"
)

type pkgRequest struct {
	Action string `json:"action"`
	Date   string `json:"date"`
}

// Validate checks if the package request fields are valid
func (r *pkgRequest) Validate() error {
	if r == nil {
		return errors.New("request is empty")
	}

	if r.Action != actionEnable && r.Action != actionDisable {
		return errors.New("bad Action value")
	}

	if _, err := time.Parse(gappsDateFormat, r.Date); err != nil {
		return errors.New("bad Date format")
	}

	return nil
}

type pkgResponse struct {
	Status string `json:"status,omitempty"`
	Error  error  `json:"error,omitempty"`
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
			resp.Error = err
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		if err := req.Validate(); err != nil {
			resp.Error = err
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		// get the release keys from the DB
		keys, err := a.db.Keys()
		if err != nil {
			resp.Error = err
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		var key string
		for i := range keys {
			if keys[i] == req.Date {
				key = req.Date
			}
		}
		if key == "" {
			resp.Error = fmt.Errorf("package with date '%s' not found", req.Date)
			respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
			return
		}

		for _, p := range gapps.PlatformValues() {
			// get the record from the DB and add it to the response
			key := getLatestArchKey(p.String(), keys)
			if key == "" {
				log.Warnf("No releases found for arch '%s'", p)
				continue
			}

			data, err := a.db.Get(key)
			if err != nil {
				resp.Error = err
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			var record db.Record
			if err = json.Unmarshal(data, &record); err != nil {
				resp.Error = err
				respondJSON(w, http.StatusInternalServerError, resp.ToJSON())
				return
			}

			// work with DB value here
		}

		switch req.Action {
		case actionEnable:
		case actionDisable:
		}
	}
}
