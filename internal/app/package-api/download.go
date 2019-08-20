package packageapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/xerrors"
)

const (
	pkgTemplate          = "open_gapps-%s-%s-%s-%s"
	reportTemplate       = "sources_report-%s-%s-%s.txt"
	masterMirrorTemplate = "https://master.dl.sourceforge.net/project/opengapps/%s/%s/%s"
	mirrorsTemplate      = "https://sourceforge.net/settings/mirror_choices?projectname=opengapps&filename=%s/%s/%s"

	zipExtension = ".zip"
	md5Extension = ".zip.md5"
	logExtension = ".versionlog.txt"
)

var templateMap = map[string]string{
	fieldZIP:          pkgTemplate + zipExtension,
	fieldZIPMirrors:   pkgTemplate + zipExtension,
	fieldMD5:          pkgTemplate + md5Extension,
	fieldVersionInfo:  pkgTemplate + logExtension,
	fieldSourceReport: reportTemplate,
}

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

		for f, t := range templateMap {
			var filename string
			switch f {
			case fieldZIPMirrors:
				filename = fmt.Sprintf(t, platform, android.HumanString(), variant, date)
				resp.SetField(f, fmt.Sprintf(mirrorsTemplate, platform, date, filename))
			case fieldSourceReport:
				filename = fmt.Sprintf(t, platform, android.HumanString(), date)
				resp.SetField(f, fmt.Sprintf(masterMirrorTemplate, platform, date, filename))
			default:
				filename = fmt.Sprintf(t, platform, android.HumanString(), variant, date)
				resp.SetField(f, fmt.Sprintf(masterMirrorTemplate, platform, date, filename))
			}
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
