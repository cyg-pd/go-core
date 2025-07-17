package k8s

import (
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/cyg-pd/go-core/httprouter"
	"github.com/gin-gonic/gin"
)

type option interface{ apply(*Provider) }
type optionFunc func(*Provider)

func (fn optionFunc) apply(cfg *Provider) { fn(cfg) }

func WithTerminationGracePeriod(period time.Duration) option {
	return optionFunc(func(p *Provider) {
		p.terminationGracePeriod = period
	})
}

func New(r httprouter.Router, opts ...option) *Provider {
	p := &Provider{
		httpRouter:             r,
		terminationGracePeriod: time.Second * 3,
	}
	for _, opt := range opts {
		opt.apply(p)
	}
	return p
}

// Provider 實作 K8s 相關功能
type Provider struct {
	httpRouter httprouter.Router

	terminationGracePeriod time.Duration

	initial atomic.Bool
	ready   atomic.Bool
}

// Boot implements provider.Provider.
func (p *Provider) Boot() error {
	if p.httpRouter == nil {
		return nil
	}

	p.ready.Store(true)

	p.httpRouter.GET("__kube/readiness", p.readiness)
	p.httpRouter.GET("__kube/liveness", p.liveness)

	return nil
}

func (p *Provider) readiness(ctx *gin.Context) {
	if !p.initial.Load() {
		p.initial.Store(true)
	}

	if p.ready.Load() {
		ctx.AbortWithStatus(http.StatusOK)
	} else {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (p *Provider) liveness(ctx *gin.Context) {
	if !p.initial.Load() {
		p.initial.Store(true)
	}

	ctx.AbortWithStatus(http.StatusOK)
}

func (p *Provider) Shutdown() error {
	if p.httpRouter == nil {
		return nil
	}

	p.ready.Store(false)

	if os.Getenv("KUBERNETES_PORT") != "" && p.initial.Load() {
		<-time.After(p.terminationGracePeriod)
	}

	return nil
}
