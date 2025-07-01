package kubernetes

import (
	"context"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/cyg-pd/go-core/httprouter"
	"github.com/cyg-pd/go-core/provider"
	"github.com/gin-gonic/gin"
)

func init() {
	provider.Register(&Provider{
		term: make(chan struct{}),
	})
}

const FailureThreshold = 5

// Provider 實作 K8s 相關功能
type Provider struct {
	initial atomic.Bool
	ready   atomic.Bool
	term    chan struct{}
}

// Boot implements provider.Provider.
func (s *Provider) Boot() error {
	s.ready.Store(true)

	r := httprouter.Default()
	r.GET("__kube/readiness", s.readiness)
	r.GET("__kube/liveness", s.liveness)

	return nil
}

func (s *Provider) readiness(ctx *gin.Context) {
	if !s.initial.Load() {
		s.initial.Store(true)
	}

	if s.ready.Load() {
		ctx.AbortWithStatus(http.StatusOK)
	} else {
		s.term <- struct{}{}
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}

func (s *Provider) liveness(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusOK)
}

func (s *Provider) Shutdown() error {
	defer close(s.term)

	if os.Getenv("KUBERNETES_PORT") == "" || !s.initial.Load() {
		return nil
	}

	s.ready.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	for range [FailureThreshold]struct{}{} {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.term:
		}
	}

	return nil
}
