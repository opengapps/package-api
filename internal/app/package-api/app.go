package packageapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/opengapps/package-api/internal/app"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"

	"github.com/google/go-github/v29/github"
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

// Application holds all the services and config
type Application struct {
	cfg    *viper.Viper
	db     *db.DB
	server *http.Server
	gh     *github.Client
}

// New creates new instance of Application
func New(cfg *viper.Viper, storage *db.DB, gh *github.Client) (*Application, error) {
	if cfg == nil {
		return nil, errors.New("passed config is nil")
	}
	if storage == nil {
		return nil, errors.New("passed storage is nil")
	}
	if gh == nil {
		return nil, errors.New("passed GitHub client is nil")
	}

	server := &http.Server{
		Addr:         cfg.GetString(config.ServerHostKey) + ":" + cfg.GetString(config.ServerPortKey),
		ReadTimeout:  cfg.GetDuration(config.HTTPTimeoutKey),
		WriteTimeout: cfg.GetDuration(config.HTTPTimeoutKey),
	}

	app.PrintInfo(cfg)
	return &Application{
		cfg:    cfg,
		db:     storage,
		gh:     gh,
		server: server,
	}, nil
}

// Run launches the Application
func (a *Application) Run() error {
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
func (a *Application) Close() error {
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
