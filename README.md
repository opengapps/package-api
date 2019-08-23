# OpenGApps Package API [![CircleCI](https://circleci.com/gh/opengapps/package-api/tree/master.svg?style=svg)](https://circleci.com/gh/opengapps/package-api/tree/master) [![Go Report Card](https://goreportcard.com/badge/github.com/opengapps/package-api)](https://goreportcard.com/report/github.com/opengapps/package-api)

This is the implementation of the [opengapps.org](https://opengapps.org) backend.

This project adheres to the Golang [project-layout](https://github.com/golang-standards/project-layout) structure.

## Purpose

Since the website has transitioned to the SourceForge storage from Github releases, we've lost Github API support.

So, to compensate for that, this service was created.

## Installation

The service uses Go modules for vendoring.

To install it, starting with Go 1.12 you can just use `go get`:

```shellscript
go get github.com/opengapps/package-api/cmd/package-api
```

Also you can clone this repo and use the build/install/run targets from [Makefile](./Makefile).

## Configuration

You can use any config file format supported by [Viper](https://github.com/spf13/viper) library (currently our preferred format is [TOML](https://github.com/toml-lang/toml)).

In case you don't want to use local config, it will rely on ENV variables with prefix `PACKAGE_API_` (can be changed in [Makefile](./Makefile)).

Example config can be found at [config_example.toml](./resources/config_example.toml).

## Usage

This API only exposes two endpoints: `/list` and `/download`.

### Request format

| Method | Endpoint    | Parameters                                                    |
| ------ | ----------- | ------------------------------------------------------------- |
| `GET`  | `/list`     | None                                                          |
| `GET`  | `/download` | `arch={ARCHITECTURE}&api={API}&variant={VARIANT}&date={DATE}` |

### Response codes

- **200**: successful response;
- **404**: on bad request format or improper parameters;
- **500**: mostly on external call failures.
