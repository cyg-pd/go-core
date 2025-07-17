package httprouter

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

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
