package packageapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/opengapps/package-api/internal/pkg/config"

	"github.com/gorilla/feeds"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"golang.org/x/xerrors"
)

const (
	feedNameTemplate = "Release notes from %s"
	feedDescTemplate = "Open GApps package release for %s architecture"

	authorName    = "opengappsbuildbot"
	copyrightText = "Copyright © 2015-2019 The Open GApps Team"
)

func (a *Application) rssHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		arch, err := validateRSSRequest(r)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		now := time.Now()
		feed := &feeds.Feed{
			Title:       fmt.Sprintf(feedNameTemplate, arch),
			Description: fmt.Sprintf(feedDescTemplate, arch),
			Link: &feeds.Link{
				Type: "application/atom+xml",
				Rel:  "self",
				Href: r.URL.Scheme + a.cfg.GetString(config.APIHostKey) + r.RequestURI,
			},
			Author:    &feeds.Author{Name: authorName},
			Created:   now,
			Copyright: copyrightText,
		}

		atom, err := feed.ToAtom()
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		respondXML(w, http.StatusOK, []byte(atom))
	}
}

func validateRSSRequest(req *http.Request) (gapps.Platform, error) {
	queryArgs := req.URL.Query()

	arch := queryArgs.Get(queryArgArch)
	if arch == "" {
		return 0, xerrors.Errorf("'%s' param is empty or missing", queryArgArch)
	}

	platform, err := gapps.PlatformString(arch)
	if err != nil {
		return 0, xerrors.Errorf("unable to parse '%s' param: '%s' is not a valid architecture", queryArgArch, arch)
	}

	return platform, nil
}
