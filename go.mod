module github.com/opengapps/package-api

go 1.13

require (
	github.com/google/go-github/v37 v37.0.0
	github.com/gorilla/feeds v1.1.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/opengapps/package-api/pkg/gapps v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	go.etcd.io/bbolt v1.3.6
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
)

replace github.com/opengapps/package-api/pkg/gapps => ./pkg/gapps
