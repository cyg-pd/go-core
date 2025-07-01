package httprouter

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type server struct {
	*gin.Engine
	server *http.Server
	config *viper.Viper
}

func createHTTPServer(host string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:    host,
		Handler: h,
	}
}

func (e *server) Run(host ...string) error {
	if len(host) == 0 {
		c := e.config

		defaultHost := c.GetString("http.host")
		if defaultHost == "" {
			return errors.New("core/httprouter: host can't be empty")
		}

		defaultPort := c.GetString("http.port")
		if defaultPort == "" {
			return errors.New("core/httprouter: port can't be empty")
		}

		host = append(host, defaultHost+":"+defaultPort)
	}

	e.server = createHTTPServer(host[0], http.AllowQuerySemicolons(e.Engine.Handler()))

	slog.Info("core/httprouter: listening and serving HTTP on " + host[0])
	if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("core/httprouter: listen error " + err.Error())
		return err
	}

	return nil
}

func (e *server) Shutdown(ctx context.Context) error {
	c := e.config
	if e.server == nil {
		return errors.New("core/httprouter: server is not running")
	}

	var cancel context.CancelFunc
	if seconds := c.GetInt("app.terminationGracePeriodSeconds"); seconds != 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(seconds)*time.Second)
		defer cancel()
	}

	if err := e.server.Shutdown(ctx); err != nil {
		slog.Error("core/httprouter: shutdown error: " + err.Error())
		return err
	}

	return nil
}

var _ Engine = (*server)(nil)
