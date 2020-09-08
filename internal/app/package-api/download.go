package packageapi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
)

const mirrorTemplate = "?r=&ts=%d&use_mirror=autoselect"

func (a *application) dlHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp models.DownloadResponse

		args, err := parseDLRequest(r)
		if err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		date := args[3]
		platform, android, variant, err := gapps.ParsePackageParts(args[:3])
		if err != nil {
			resp.Error = err.Error()
			respondJSON(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		now := time.Now().Unix()
		for f := range models.TemplateMap {
			url := models.NewDownloadLink(f, date, platform, android, variant)
			if f != models.FieldZIPMirrors {
				url += fmt.Sprintf(mirrorTemplate, now)
			}
			resp.SetField(f, url)
		}

		respondJSON(w, http.StatusOK, resp.ToJSON())
	}
}

func parseDLRequest(req *http.Request) ([]string, error) {
	queryArgs := req.URL.Query()

	arch := queryArgs.Get(queryArgArch)
	if arch == "" {
		return nil, fmt.Errorf(missingParamErrTemplate, queryArgArch)
	}

	api := queryArgs.Get(queryArgAPI)
	if api == "" {
		return nil, fmt.Errorf(missingParamErrTemplate, queryArgAPI)
	}
	api = strings.Replace(api, ".", "", 1)

	variant := queryArgs.Get(queryArgVariant)
	if variant == "" {
		return nil, fmt.Errorf(missingParamErrTemplate, queryArgVariant)
	}

	date := queryArgs.Get(queryArgDate)
	if date == "" {
		return nil, fmt.Errorf(missingParamErrTemplate, queryArgDate)
	}

	return []string{arch, api, variant, date}, nil
}
