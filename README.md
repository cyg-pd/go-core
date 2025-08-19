# go-core

[![tag](https://img.shields.io/github/tag/cyg-pd/go-core.svg)](https://github.com/cyg-pd/go-core/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.24-%23007d9c)
[![GoDoc](https://godoc.org/github.com/cyg-pd/go-core?status.svg)](https://pkg.go.dev/github.com/cyg-pd/go-core)
![Build Status](https://github.com/cyg-pd/go-core/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/cyg-pd/go-core)](https://goreportcard.com/report/github.com/cyg-pd/go-core)
[![Coverage](https://img.shields.io/codecov/c/github/cyg-pd/go-core)](https://codecov.io/gh/cyg-pd/go-core)
[![Contributors](https://img.shields.io/github/contributors/cyg-pd/go-core)](https://github.com/cyg-pd/go-core/graphs/contributors)
[![License](https://img.shields.io/github/license/cyg-pd/go-core)](./LICENSE)

## 🚀 Install

```sh
go get github.com/cyg-pd/go-core@v1
```

This library is v1 and follows SemVer strictly.

No breaking changes will be made to exported APIs before v2.0.0.

## 💡 Usage

You can import `core` using:

```go
package main

import (
	"context"
	"log/slog"

	"github.com/cyg-pd/go-core/application"
	"github.com/cyg-pd/go-core/config"
	"github.com/cyg-pd/go-core/httprouter"
	"github.com/cyg-pd/go-core/httprouter/middleware"
	"github.com/cyg-pd/go-core/logger"
	"github.com/cyg-pd/go-core/msgrouter"
	"github.com/cyg-pd/go-core/provider"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

var app *application.Application

func init() {
	f := pflag.CommandLine
	v := config.New()

	httprouter.SetupFlags(f, v, "http")
	logger.SetupFlags(f, v)

	pflag.Parse()
	slog.SetDefault(logger.New(v))

	app = application.New(httprouter.New(v), msgrouter.New())
	app.AddBeforeRunHook(func(ctx context.Context) error { return provider.Boot() })
	app.AddBeforeShutdownHook(func(ctx context.Context) error { return provider.Shutdown() })
}

func main() {
	ctx := context.Background()
	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}

// http middleware
func init() {
	r := app.HTTPRouter()

	r.Use(gin.Recovery())
	r.Use(middleware.Hostname())
	r.Use(middleware.CorrelationID())
}

```
