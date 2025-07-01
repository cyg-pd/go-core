package logger

import (
	"log/slog"

	coreconfig "github.com/cyg-pd/go-core/config"
	"github.com/cyg-pd/go-slogx"
	_ "github.com/cyg-pd/go-slogx/driver/console"
	_ "github.com/cyg-pd/go-slogx/driver/json"
	_ "github.com/cyg-pd/go-slogx/driver/otel"
	_ "github.com/cyg-pd/go-slogx/driver/text"
	"github.com/spf13/pflag"
)

func init() {
	pflag.String("log-driver", "console", "Log driver")
	pflag.String("log-level", "info", "Log level (debug, info, warn, error)")
	pflag.String("log-opts", "{}", "Log options")
	pflag.Bool("log-source", false, "Log source")

	coreconfig.BindPFlag("log.driver", pflag.Lookup("log-driver"))
	coreconfig.BindPFlag("log.level", pflag.Lookup("log-level"))
	coreconfig.BindPFlag("log.opts", pflag.Lookup("log-opts"))
	coreconfig.BindPFlag("log.source", pflag.Lookup("log-source"))
}

func Parse() {
	var c struct {
		Log struct {
			Driver  string `mapstructure:"driver"`
			Level   string `mapstructure:"level"`
			Source  bool   `mapstructure:"source"`
			Options string `mapstructure:"opts"`
		} `mapstructure:"log"`
	}

	if err := coreconfig.Unmarshal(&c); err != nil {
		panic(err)
	}

	slog.SetDefault(slogx.New(
		slogx.WithDriver(c.Log.Driver, c.Log.Options),
		slogx.WithLevel(c.Log.Level),
		slogx.WithSource(c.Log.Source),
	))
}
