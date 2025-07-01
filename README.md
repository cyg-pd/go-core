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
	"net/http"
	"strings"

	"github.com/cyg-pd/go-core/application"
	"github.com/cyg-pd/go-core/config"
	"github.com/cyg-pd/go-core/httprouter"
	"github.com/cyg-pd/go-core/httprouter/middleware"
	"github.com/cyg-pd/go-core/logger"
	"github.com/cyg-pd/go-otelx"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

var app = application.Default()

func init() {
	config.Parse()
	logger.Parse()
}

func main() {
	ctx := context.Background()

	app.Run(ctx)
}

// http middleware
func init() {
	r := httprouter.Default()

	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware(otelx.ServiceName(), otelgin.WithFilter(
		RequestDoubleUnderscore,
	)))
	r.Use(middleware.Hostname())
	r.Use(middleware.CorrelationID())
}

func RequestDoubleUnderscore(r *http.Request) bool {
	return !strings.Contains(r.URL.Path, "__")
}
```
