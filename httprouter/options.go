package httprouter

import (
	"github.com/gin-gonic/gin"
)

type option interface{ apply(*server) }
type optionFunc func(*server)

func (fn optionFunc) apply(cfg *server) { fn(cfg) }

func WithEngine(engine *gin.Engine) option {
	return optionFunc(func(r *server) {
		r.Engine = engine
	})
}
