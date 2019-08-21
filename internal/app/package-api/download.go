package packageapi

import (
	"net/http"
	"strings"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/xerrors"
)

func (a *Application) dlHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var resp DownloadResponse

		args, err := validateDLRequest(r)
		if err != nil {
			resp.Error = err.Error()
			respond(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		date := args[3]
		platform, android, variant, err := gapps.ParsePackageParts(args[:3])
		if err != nil {
			resp.Error = err.Error()
			respond(w, http.StatusBadRequest, resp.ToJSON())
			return
		}

		for f := range templateMap {
			resp.SetField(f, formatLink(f, date, platform, android, variant))
		}

		respond(w, http.StatusOK, resp.ToJSON())
	}
}

func validateDLRequest(req *http.Request) ([]string, error) {
	queryArgs := req.URL.Query()

	arch := queryArgs.Get("arch")
	if arch == "" {
		return nil, xerrors.New("'arch' param is empty or missing")
	}

	api := queryArgs.Get("api")
	if api == "" {
		return nil, xerrors.New("'api' param is empty or missing")
	}
	api = strings.Replace(api, ".", "", 1)

	variant := queryArgs.Get("variant")
	if variant == "" {
		return nil, xerrors.New("'variant' param is empty or missing")
	}

	date := queryArgs.Get("date")
	if date == "" {
		return nil, xerrors.New("'date' param is empty or missing")
	}

	return []string{arch, api, variant, date}, nil
}
