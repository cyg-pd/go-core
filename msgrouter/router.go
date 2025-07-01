package msgrouter

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/cyg-pd/go-core/internal/utils"
	"github.com/cyg-pd/go-otelx"
	"github.com/cyg-pd/go-watermillx/cqrx"
	"github.com/cyg-pd/go-watermillx/manager"
	"github.com/cyg-pd/go-watermillx/opentelemetry"
	"github.com/cyg-pd/go-watermillx/opentelemetry/metrics"
)

var defaultRouter atomic.Pointer[Router]

func init() {
	defaultRouter.Store(New())
}

// Default returns the default [Router].
func Default() *Router {
	return defaultRouter.Load()
}

// SetDefault makes app the default [Router]
func SetDefault(app *Router) {
	defaultRouter.Store(app)
}

type Router struct {
	*message.Router

	once    sync.Once
	manager *manager.Manager

	metricBuilder metrics.Builder
}

func (r *Router) Init() {
	r.once.Do(func() {
		var err error
		r.Router, err = message.NewRouter(message.RouterConfig{}, watermill.NewSlogLogger(nil))
		if err != nil {
			panic(err)
		}

		r.AddMiddleware(middleware.Recoverer)

		r.metricBuilder = metrics.NewOpenTelemetryMetricsBuilder(otelx.Meter(), "", "")
		r.AddMiddleware(opentelemetry.HandlerMiddleware(r.metricBuilder))
		r.AddPublisherDecorators(opentelemetry.PublisherDecorator(r.metricBuilder))
		r.AddSubscriberDecorators(opentelemetry.SubscriberDecorator(r.metricBuilder))

		r.manager = manager.New(r.Router)
	})
}

func (r *Router) MustUse(name ...string) *cqrx.CQRS { return utils.Must(r.Use(name...)) }
func (r *Router) Use(name ...string) (*cqrx.CQRS, error) {
	r.Init()

	n := "default"
	if len(name) > 0 {
		n = name[0]
	}

	cqrs, err := r.manager.UseCQRS(n)
	if err != nil {
		return nil, err
	}

	cqrs.AddPublisherDecorator(opentelemetry.PublisherDecorator(r.metricBuilder))
	cqrs.AddSubscriberDecorator(opentelemetry.SubscriberDecorator(r.metricBuilder))

	return cqrs, nil
}

func (r *Router) Run(ctx context.Context) error {
	r.Init()
	return r.Router.Run(ctx)
}

func (r *Router) Shutdown(ctx context.Context) error {
	r.Init()
	return r.Router.Close()
}

func New() *Router { return &Router{} }
