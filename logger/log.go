package logger

import (
	"log/slog"

	"github.com/cyg-pd/go-slogx"
	_ "github.com/cyg-pd/go-slogx/driver/console"
	_ "github.com/cyg-pd/go-slogx/driver/json"
	_ "github.com/cyg-pd/go-slogx/driver/otel"
	_ "github.com/cyg-pd/go-slogx/driver/text"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetupFlags(f *pflag.FlagSet, v *viper.Viper) {
	f.String("log-driver", "console", "Log driver")
	f.String("log-level", "info", "Log level (debug, info, warn, error)")
	f.String("log-opts", "{}", "Log options")
	f.Bool("log-source", false, "Log source")

	_ = v.BindPFlag("log.driver", f.Lookup("log-driver"))
	_ = v.BindPFlag("log.level", f.Lookup("log-level"))
	_ = v.BindPFlag("log.opts", f.Lookup("log-opts"))
	_ = v.BindPFlag("log.source", f.Lookup("log-source"))
}

func Parse(v *viper.Viper) {
	var c struct {
		Log struct {
			Driver  string `mapstructure:"driver"`
			Level   string `mapstructure:"level"`
			Source  bool   `mapstructure:"source"`
			Options string `mapstructure:"opts"`
		} `mapstructure:"log"`
	}

	if err := v.Unmarshal(&c); err != nil {
		panic(err)
	}

	if len(c.Log.Driver) == 0 {
		c.Log.Driver = "console"
	}

	if len(c.Log.Level) == 0 {
		c.Log.Level = "error"
	}

	slog.SetDefault(slogx.New(
		slogx.WithDriver(c.Log.Driver, c.Log.Options),
		slogx.WithLevel(c.Log.Level),
		slogx.WithSource(c.Log.Source),
	))
}
