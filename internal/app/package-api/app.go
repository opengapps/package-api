package packageapi

import (
	"context"
	"net/http"

	"github.com/opengapps/package-api/internal/app"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/opengapps/package-api/internal/pkg/db"

	"github.com/google/go-github/v28/github"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
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
		Addr: cfg.GetString(config.ServerHostKey) + ":" + cfg.GetString(config.ServerPortKey),
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
	r := mux.NewRouter()
	r.HandleFunc(a.cfg.GetString(config.DownloadEndpointKey), a.dlHandler()).
		Host(a.cfg.GetString(config.APIHostKey)).
		Methods(http.MethodGet).
		Queries(queryArgArch, "", queryArgAPI, "", queryArgVariant, "", queryArgDate, "")
	r.HandleFunc(a.cfg.GetString(config.ListEndpointKey), a.listHandler()).
		Host(a.cfg.GetString(config.APIHostKey)).
		Methods(http.MethodGet)
	r.HandleFunc(a.cfg.GetString(config.RSSEndpointKey), a.rssHandler()).
		Host(a.cfg.GetString(config.APIHostKey)).
		Methods(http.MethodGet).
		Queries(queryArgArch, "")
	a.server.Handler = handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
	)(r)

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
