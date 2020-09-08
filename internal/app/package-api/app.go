package packageapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/opengapps/package-api/internal/app"
	"github.com/opengapps/package-api/internal/pkg/config"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	missingParamErrTemplate = "'%s' param is empty or missing"

	queryArgArch    = "arch"
	queryArgAPI     = "api"
	queryArgVariant = "variant"
	queryArgDate    = "date"
)

type application struct {
	cfg     *viper.Viper
	server  *http.Server
	storage Storage
}

// New creates new instance of Application
func New(opts ...Option) (*application, error) {
	a := &application{}
	for _, opt := range opts {
		if err := opt(a); err != nil {
			return nil, fmt.Errorf("unable to create client: %w", err)
		}
	}
	if a.cfg == nil {
		return nil, errors.New("passed config is nil")
	}
	if a.storage == nil {
		return nil, errors.New("passed storage is nil")
	}

	a.server = &http.Server{
		Addr:         a.cfg.GetString(config.ServerHostKey) + ":" + a.cfg.GetString(config.ServerPortKey),
		ReadTimeout:  a.cfg.GetDuration(config.HTTPTimeoutKey),
		WriteTimeout: a.cfg.GetDuration(config.HTTPTimeoutKey),
	}
	app.PrintInfo(a.cfg)

	return a, nil
}

// Run launches the Application
func (a *application) Run() error {
	// init router
	r := mux.NewRouter().
		Host(a.cfg.GetString(config.APIHostKey)).
		Subrouter()

	// set normal handlers
	r.Name("download").Path(a.cfg.GetString(config.DownloadEndpointKey)).
		Methods(http.MethodGet).
		Queries(queryArgArch, "", queryArgAPI, "", queryArgVariant, "", queryArgDate, "").
		HandlerFunc(a.dlHandler())
	r.Name("list").Path(a.cfg.GetString(config.ListEndpointKey)).
		Methods(http.MethodGet).
		HandlerFunc(a.listHandler())
	r.Name("rss").Path(a.cfg.GetString(config.RSSEndpointKey)).
		Methods(http.MethodGet).
		HandlerFunc(a.rssHandler())

	// set auth-covered handlers
	r.Name("pkg").Path(a.cfg.GetString(config.PkgEndpointKey)).
		Methods(http.MethodPost).
		Handler(authMiddleware(a.cfg.GetString(config.AuthKey), a.pkgHandler()))

	// set handler with middlewares
	a.server.Handler = withMiddlewares(r)

	// serve
	log.Warnf("Serving at %s", a.server.Addr)
	return a.server.ListenAndServe()
}

// Close stops the Application
func (a *application) Close() error {
	return a.server.Shutdown(context.Background())
}

func respondJSON(w http.ResponseWriter, code int, body []byte) {
	respond(w, "application/json; charset=utf-8", code, body)
}

func respondXML(w http.ResponseWriter, code int, body []byte) {
	respond(w, "application/atom+xml; charset=utf-8", code, body)
}

func respond(w http.ResponseWriter, contentType string, code int, body []byte) {
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	if _, err := w.Write(body); err != nil {
		log.WithError(err).Error("Unable to write answer")
	}
}

func errToBytes(err error) []byte {
	if err == nil {
		return nil
	}
	return []byte(err.Error())
}
