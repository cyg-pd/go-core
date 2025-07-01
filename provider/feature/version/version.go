package pprof

import (
	"encoding/json"

	"github.com/cyg-pd/go-core/application"
	"github.com/cyg-pd/go-core/provider"
	"github.com/gin-gonic/gin"
)

func init() {
	provider.Register(&VersionProvider{})
}

type VersionProvider struct{ provider.NoopProvider }

func (*VersionProvider) Boot() error {
	app := application.Default()
	v := app.Version
	if v == nil {
		return nil
	}

	res, err := json.Marshal(map[string]any{
		"version":    v.Version,
		"build_user": v.BuildUser,
		"build_time": v.BuildTime,
	})
	if err != nil {
		return err
	}

	r := app.HTTPRouter()
	r.GET("__version", func(c *gin.Context) {
		c.Data(200, "application/json; charset=utf-8", res)
	})

	return nil
}

var _ provider.Provider = (*VersionProvider)(nil)
