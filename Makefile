CMD:=package-api
PKG_LIST:=$(shell go list ./...)
GIT_HASH?=$(shell git log --format="%h" -n 1 2> /dev/null)
GIT_BRANCH?=$(shell git branch 2> /dev/null | grep '*' | cut -f2 -d' ')
GIT_TAG:=$(shell git describe --exact-match --abbrev=0 --tags 2> /dev/null)
APP_VERSION?=$(if $(GIT_TAG),$(GIT_TAG),$(shell git describe --all --long HEAD 2> /dev/null))
GO_VERSION:=$(shell go version)
GO_VERSION_SHORT:=$(shell echo $(GO_VERSION)|sed -E 's/.* go(.*) .*/\1/g')

export GOPROXY=https://proxy.golang.org
BUILD_ENVPARMS:=CGO_ENABLED=0
BUILD_TS:=$(shell date +%FT%T%z)
LDFLAGS:=-X 'github.com/opengapps/$(CMD)/internal/app.Name=$(CMD)'\
		 -X 'github.com/opengapps/$(CMD)/internal/app.Version=$(APP_VERSION)'\
         -X 'github.com/opengapps/$(CMD)/internal/app.BuildTS=$(BUILD_TS)'\
         -X 'github.com/opengapps/$(CMD)/internal/app.GoVersion=$(GO_VERSION_SHORT)'\
         -X 'github.com/opengapps/$(CMD)/internal/app.GitHash=$(GIT_HASH)'\
         -X 'github.com/opengapps/$(CMD)/internal/app.GitBranch=$(GIT_BRANCH)'

# install project dependencies
.PHONY: deps
deps:
	$(info #Install dependencies and clean up...)
	go mod tidy

# run all tests
.PHONY: test
test:
	$(info #Running tests...)
	go test -v -cover -race ./...

# run all tests with coverage
.PHONY: test-cover
test-cover:
	$(info #Running tests with coverage...)
	go test -v -coverprofile=coverage.out -race $(PKG_LIST)
	@go tool cover -func=coverage.out | grep total
	@rm -f coverage.out
	
.PHONY: fast-build
fast-build: deps
	$(info #Building binary...)
	$(shell $(BUILD_ENVPARMS) go build -ldflags "-s -w $(LDFLAGS)" -o ./bin/$(CMD) ./cmd/$(CMD))
	@echo

.PHONY: build
build: deps fast-build test

.PHONY: install
install:
	$(info #Installing binary...)
	$(shell $(BUILD_ENVPARMS) go install -ldflags "-s -w $(LDFLAGS)" ./cmd/$(CMD))
	@echo

.PHONY: run
run:
	$(info #Running the main program...)
	$(shell $(BUILD_ENVPARMS) go run -ldflags "-s -w $(LDFLAGS)" ./cmd/$(CMD))

.PHONY: lint
lint:
	golangci-lint run --config=.golangci.yml --timeout=5m ./...
