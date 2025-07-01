package application

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/cyg-pd/go-core/httprouter"
	"github.com/cyg-pd/go-core/msgrouter"
	_ "github.com/cyg-pd/go-otelx/autoconf"
	"golang.org/x/sync/errgroup"
)

var defaultApplication atomic.Pointer[Application]

func init() {
	defaultApplication.Store(New(
		httprouter.Default(),
		msgrouter.Default(),
	))
}

// Default returns the default [Application].
func Default() *Application { return defaultApplication.Load() }

// SetDefault makes app the default [Application]
func SetDefault(app *Application) { defaultApplication.Store(app) }

type Version struct {
	Version   string
	BuildTime string
	BuildUser string
}

type Application struct {
	Version *Version

	msgChan  chan struct{}
	httpChan chan struct{}

	msgRouter  *msgrouter.Router
	httpRouter *httprouter.Server

	beforeRunHooks      []func(ctx context.Context) error
	beforeShutdownHooks []func(ctx context.Context) error
}

func (a *Application) HTTPRouter() *httprouter.Server   { return a.httpRouter }
func (a *Application) MessageRouter() *msgrouter.Router { return a.msgRouter }

func (a *Application) AddBeforeRunHook(hook func(ctx context.Context) error) {
	a.beforeRunHooks = append(a.beforeRunHooks, hook)
}

func (a *Application) AddBeforeShutdownHook(hook func(ctx context.Context) error) {
	a.beforeShutdownHooks = append(a.beforeShutdownHooks, hook)
}

func (a *Application) SetVersion(version, buildUser, buildTime string) {
	a.Version = &Version{
		Version:   version,
		BuildTime: buildTime,
		BuildUser: buildUser,
	}
}

func (a *Application) Run(ctx context.Context) error {
	for _, hook := range a.beforeRunHooks {
		if err := hook(ctx); err != nil {
			return err
		}
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Graceful shutdown
	errGrp, ctx := errgroup.WithContext(ctx)
	a.runEventBus(errGrp)
	a.runHTTP(errGrp)

	// Shutdown application
	errGrp.Go(func() error {
		r := a.HTTPRouter()
		ev := a.MessageRouter()

		if r == nil && ev == nil {
			return errors.New("core/application: http/event router is nil")
		}

		<-ctx.Done()
		defer func() {
			if r == nil {
				close(a.msgChan)
				return
			}

			close(a.httpChan)
		}()

		for _, hook := range a.beforeShutdownHooks {
			if err := hook(ctx); err != nil {
				return err
			}
		}

		return nil
	})

	// Force shutdown
	go a.forceShutdown(ctx)

	if err := errGrp.Wait(); err != nil {
		slog.Error("core/application: " + err.Error())
	}

	slog.Info("core/application: server stopped")

	return nil
}

func (a *Application) runHTTP(errGrp *errgroup.Group) {
	r := a.HTTPRouter()
	if r == nil {
		return
	}

	// Start http server
	errGrp.Go(func() error { return r.Run() })

	// Stop http server
	errGrp.Go(func() error {
		<-a.httpChan
		defer close(a.msgChan)
		return r.Shutdown(context.Background())
	})
}

func (a *Application) runEventBus(errGrp *errgroup.Group) {
	ev := a.MessageRouter()
	if ev == nil {
		return
	}

	// Start event
	errGrp.Go(func() error { return ev.Run(context.Background()) })

	// Stop event
	errGrp.Go(func() error {
		<-a.msgChan
		return ev.Shutdown(context.Background())
	})
}

func (a *Application) forceShutdown(ctx context.Context) {
	slog.Info("core/application: press Ctrl+C to stop server")
	<-ctx.Done()

	fCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	slog.Info("core/application: shutting down gracefully, press Ctrl+C again to force stop")
	<-fCtx.Done()

	slog.Info("core/application: shutting down immediately")
	stop()
	os.Exit(1)
}

func New(httpRouter *httprouter.Server, msgRouter *msgrouter.Router, opts ...option) *Application {
	app := &Application{
		msgChan:    make(chan struct{}),
		httpChan:   make(chan struct{}),
		msgRouter:  msgRouter,
		httpRouter: httpRouter,
	}
	for _, opt := range opts {
		opt.apply(app)
	}
	return app
}
