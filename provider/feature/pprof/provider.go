package pprof

import (
	"github.com/cyg-pd/go-core/httprouter"
	"github.com/cyg-pd/go-core/provider"
	"github.com/gin-contrib/pprof"
)

func init() {
	provider.Register(&PprofProvider{})
}

type PprofProvider struct{ provider.NoopProvider }

func (*PprofProvider) Boot() error {
	pprof.Register(httprouter.Default(), "__debug/pprof")
	return nil
}

var _ provider.Provider = (*PprofProvider)(nil)
