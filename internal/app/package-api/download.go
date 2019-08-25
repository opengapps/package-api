package packageapi

import (
	"net/http"
	"strings"

	"github.com/opengapps/package-api/internal/pkg/link"
	"github.com/opengapps/package-api/internal/pkg/models"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/xerrors"
)

const (
	queryArgArch    = "arch"
	queryArgAPI     = "api"
	queryArgVariant = "variant"
	queryArgDate    = "date"
)

func (a *Application) dlHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp models.DownloadResponse

		args, err := validateDLRequest(r)
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

		for f := range link.TemplateMap {
			resp.SetField(f, link.New(f, date, platform, android, variant))
		}

		respondJSON(w, http.StatusOK, resp.ToJSON())
	}
}

func validateDLRequest(req *http.Request) ([]string, error) {
	queryArgs := req.URL.Query()

	arch := queryArgs.Get(queryArgArch)
	if arch == "" {
		return nil, xerrors.Errorf("'%s' param is empty or missing", queryArgArch)
	}

	api := queryArgs.Get(queryArgAPI)
	if api == "" {
		return nil, xerrors.Errorf("'%s' param is empty or missing", queryArgAPI)
	}
	api = strings.Replace(api, ".", "", 1)

	variant := queryArgs.Get(queryArgVariant)
	if variant == "" {
		return nil, xerrors.Errorf("'%s' param is empty or missing", queryArgVariant)
	}

	date := queryArgs.Get(queryArgDate)
	if date == "" {
		return nil, xerrors.Errorf("'%s' param is empty or missing", queryArgDate)
	}

	return []string{arch, api, variant, date}, nil
}
