package packageapi

import (
	"context"
	"net/http"

	"github.com/opengapps/package-api/internal/app"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"

	"github.com/google/go-github/v28/github"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
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
		return nil, xerrors.New("config is nil")
	}
	if storage == nil {
		return nil, xerrors.New("storage is nil")
	}
	if gh == nil {
		return nil, xerrors.New("GitHub client is nil")
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
		Methods(http.MethodGet).
		Subrouter()

	r.Name("download").Path(a.cfg.GetString(config.DownloadEndpointKey)).
		Queries(queryArgArch, "", queryArgAPI, "", queryArgVariant, "", queryArgDate, "").
		HandlerFunc(a.dlHandler())
	r.Name("list").Path(a.cfg.GetString(config.ListEndpointKey)).
		HandlerFunc(a.listHandler())
	r.Name("rss").Path(a.cfg.GetString(config.RSSEndpointKey)).
		HandlerFunc(a.rssHandler())

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
		w.Header().Add("Content-Type", contentType)
	}
	w.WriteHeader(code)
	_, err := w.Write(body)
	if err != nil {
		log.WithError(err).Error("Unable to write answer")
	}
}
