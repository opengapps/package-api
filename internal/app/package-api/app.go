package packageapi

import (
	"context"
	"log"
	"net/http"

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
	m := http.NewServeMux()
	m.HandleFunc(a.cfg.GetString(config.DownloadEndpointKey), a.dlHandler())
	m.HandleFunc(a.cfg.GetString(config.ListEndpointKey), a.listHandler())
	a.server.Handler = m

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
