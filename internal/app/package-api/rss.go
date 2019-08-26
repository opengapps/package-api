package packageapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
	"github.com/opengapps/package-api/internal/pkg/config"
	"golang.org/x/xerrors"
)

const (
	feedNameTemplate = "Release notes from %s"
	feedDescTemplate = "Open GApps package release for %s architecture"

	authorName    = "opengappsbuildbot"
	copyrightText = "Copyright © 2015-2019 The Open GApps Team"

	creationTS = 1566816430 // time of the feed creation
)

var baseFeed = feeds.Feed{
	Link: &feeds.Link{
		Type: "application/atom+xml",
		Rel:  "self",
	},
	Author:    &feeds.Author{Name: authorName},
	Created:   time.Unix(creationTS, 0),
	Copyright: copyrightText,
}

func (a *Application) rssHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		feed := &feeds.Feed{}
		*feed = baseFeed

		arch, err := parseRSSRequest(r)
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		feed.Title = fmt.Sprintf(feedNameTemplate, arch)
		feed.Description = fmt.Sprintf(feedDescTemplate, arch)
		feed.Link.Href = r.URL.Scheme + a.cfg.GetString(config.APIHostKey) + r.RequestURI

		atom, err := feed.ToAtom()
		if err != nil {
			respond(w, "", http.StatusInternalServerError, []byte(err.Error()))
			return
		}

		respondXML(w, http.StatusOK, []byte(atom))
	}
}

func parseRSSRequest(req *http.Request) (gapps.Platform, error) {
	arch, ok := mux.Vars(req)[queryArgArch]
	if !ok {
		return 0, xerrors.Errorf(missingParamErrTemplate, queryArgArch)
	}

	platform, err := gapps.PlatformString(arch)
	if err != nil {
		return 0, xerrors.Errorf("unable to parse '%s' param: '%s' is not a valid architecture", queryArgArch, arch)
	}

	return platform, nil
}
