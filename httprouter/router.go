package httprouter

import (
	"context"
	"sync/atomic"

	"github.com/cyg-pd/go-core/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var defaultRouter atomic.Pointer[Server]

func init() {
	defaultRouter.Store(New(config.Default()))
}

// Default returns the default [Server].
func Default() *Server {
	return defaultRouter.Load()
}

// SetDefault makes app the default [Server]
func SetDefault(app *Server) {
	defaultRouter.Store(app)
}

type Router = gin.IRouter

type Engine interface {
	Router

	NoRoute(handlers ...gin.HandlerFunc)
	Run(host ...string) error
	Shutdown(ctx context.Context) error
}

type Server struct{ Engine }

func New(config *viper.Viper, opts ...option) *Server {
	r := &server{
		Engine: gin.New(),
		config: config,
	}
	for _, opt := range opts {
		opt.apply(r)
	}
	return &Server{r}
}
