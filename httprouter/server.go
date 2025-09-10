package httprouter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type server struct {
	*gin.Engine
	server *http.Server
	config *viper.Viper
}

func (e *server) createHTTPServer(host string, h http.Handler) *http.Server {
	s := &http.Server{
		Addr:                         host,
		Handler:                      h,
		DisableGeneralOptionsHandler: true,
		ReadHeaderTimeout:            10 * time.Second,
		ReadTimeout:                  60 * time.Second,
		WriteTimeout:                 60 * time.Second,
		IdleTimeout:                  120 * time.Second,
	}

	if t := e.config.GetDuration("http.server.idle.timeout"); t > 0 {
		slog.Debug("core/httprouter: http.server.idle.timeout is set to " + t.String())
		s.IdleTimeout = t
	}

	if t := e.config.GetDuration("http.server.write.timeout"); t > 0 {
		slog.Debug("core/httprouter: http.server.write.timeout is set to " + t.String())

		s.WriteTimeout = t
	}

	if t := e.config.GetDuration("http.server.read.timeout"); t > 0 {
		slog.Debug("core/httprouter: http.server.read.timeout is set to " + t.String())
		s.ReadTimeout = t
	}

	return s
}

func (e *server) Run(host ...string) error {
	if len(host) == 0 {
		host = append(
			host,
			e.config.GetString("http.host")+":"+strconv.Itoa(e.config.GetInt("http.port")),
		)
	}

	ln, err := net.Listen("tcp", host[0])
	if err != nil {
		return fmt.Errorf("core/httprouter: %w", err)
	}

	e.server = e.createHTTPServer(ln.Addr().String(), http.AllowQuerySemicolons(e.Handler()))

	slog.Info("core/httprouter: serving HTTP on " + ln.Addr().String())
	if err := e.server.Serve(ln); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("core/httprouter: %w", err)
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
		return fmt.Errorf("core/httprouter: %w", err)
	}

	return nil
}

var _ Engine = (*server)(nil)
