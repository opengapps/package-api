package packageapi

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/opengapps/package-api/internal/app"
	"github.com/opengapps/package-api/internal/pkg/cache"
	"github.com/opengapps/package-api/internal/pkg/config"
	"github.com/spf13/viper"
	"golang.org/x/xerrors"
)

// Application holds all the services and config
type Application struct {
	cfg    *viper.Viper
	cache  *cache.Cache
	server *http.Server
}

// New creates new instance of Application
func New(cfg *viper.Viper, cache *cache.Cache) (*Application, error) {
	if cfg == nil {
		return nil, xerrors.New("config is nil")
	}
	if cache == nil {
		return nil, xerrors.New("cache is nil")
	}

	server := &http.Server{
		Addr: cfg.GetString(config.ServerHostKey) + ":" + cfg.GetString(config.ServerPortKey),
	}

	app.PrintInfo(cfg)
	return &Application{
		cfg:    cfg,
		cache:  cache,
		server: server,
	}, nil
}

// Run launches the Application
func (a *Application) Run() error {
	r := mux.NewRouter()
	r.HandleFunc(a.cfg.GetString(config.DownloadEndpointKey), a.dlHandler()).
		Host(a.cfg.GetString(config.ServerHostKey)).
		Methods(http.MethodGet)
	r.HandleFunc(a.cfg.GetString(config.ListEndpointKey), a.listHandler()).
		Host(a.cfg.GetString(config.ServerHostKey)).
		Methods(http.MethodGet)
	a.server.Handler = handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
	)(r)

	log.Printf("Serving at %s", a.server.Addr)
	return a.server.ListenAndServe()
}

// Close stops the Application
func (a *Application) Close() error {
	a.cache.Clear()
	return a.server.Shutdown(context.Background())
}

func respond(w http.ResponseWriter, code int, body []byte) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(body)
	if err != nil {
		log.Println("Unable to write answer:", err)
	}
}
