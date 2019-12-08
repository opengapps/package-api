package packageapi

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

const (
	authHeader = "Authorization"
	authFormat = "Bearer %s"
)

func withMiddlewares(next http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"*"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost}),
	)(
		handlers.CustomLoggingHandler(
			os.Stdout,
			handlers.RecoveryHandler(
				handlers.RecoveryLogger(log.StandardLogger()),
				handlers.PrintRecoveryStack(true),
			)(next),
			logFormatter,
		),
	)
}

// authMiddleware checks basic authentication header
func authMiddleware(authKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.Header.Get(authHeader), fmt.Sprintf(authFormat, authKey)) {
			respond(w, "", http.StatusUnauthorized, nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// logFormatter impelements handlers.LogFormatter
func logFormatter(_ io.Writer, params handlers.LogFormatterParams) {
	fields := log.Fields{
		"Proto":  params.Request.Proto,
		"Method": params.Request.Method,
		"URL":    params.Request.RequestURI,
		"Code":   params.StatusCode,
		"Size":   params.Size,
	}

	host, _, err := net.SplitHostPort(params.Request.RemoteAddr)
	if err != nil {
		host = params.Request.RemoteAddr
	}
	fields["Host"] = host

	status := ""
	if params.StatusCode != http.StatusOK {
		status = http.StatusText(params.StatusCode)
	}
	log.WithFields(fields).Debug(status)
}
