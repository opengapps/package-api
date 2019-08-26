package packageapi

import (
	"io"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func withMiddlewares(h http.Handler) http.Handler {
	return handlers.CORS(
		handlers.AllowedHeaders([]string{"*"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{http.MethodGet}),
	)(
		handlers.CustomLoggingHandler(
			os.Stdout,
			handlers.RecoveryHandler(
				handlers.RecoveryLogger(log.StandardLogger()),
				handlers.PrintRecoveryStack(true),
			)(h),
			logFormatter,
		),
	)
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
	log.WithFields(fields).Info(status)
}
