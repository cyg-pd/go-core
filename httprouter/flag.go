package httprouter

import (
	"github.com/cyg-pd/go-kebabcase"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetupFlags(f *pflag.FlagSet, v *viper.Viper, prefix string) {
	dashHost, dotHost := buildKey(prefix, "host")
	dashPort, dotPort := buildKey(prefix, "port")
	dashIdleTimeout, dotIdleTimeout := buildKey(prefix, "server.idleTimeout")
	dashWriteTimeout, dotWriteTimeout := buildKey(prefix, "server.writeTimeout")
	dashReadTimeout, dotReadTimeout := buildKey(prefix, "server.readTimeout")
	dashReadHeaderTimeout, dotReadHeaderTimeout := buildKey(prefix, "server.readHeaderTimeout")

	f.String(dashHost, "", "HTTP Server listen host (e.g. 127.0.0.1)")
	f.Uint(dashPort, 0, "HTTP Server listen port (e.g. 8080)")
	f.Duration(dashIdleTimeout, 0, "HTTP Server Idle Timeout")
	f.Duration(dashWriteTimeout, 0, "HTTP Server Write Timeout")
	f.Duration(dashReadTimeout, 0, "HTTP Server Read Timeout")
	f.Duration(dashReadHeaderTimeout, 0, "HTTP Server Read Header Timeout")

	_ = v.BindPFlag(dotHost, f.Lookup(dashHost))
	_ = v.BindPFlag(dotPort, f.Lookup(dashPort))
	_ = v.BindPFlag(dotIdleTimeout, f.Lookup(dashIdleTimeout))
	_ = v.BindPFlag(dotWriteTimeout, f.Lookup(dashWriteTimeout))
	_ = v.BindPFlag(dotReadTimeout, f.Lookup(dashReadTimeout))
	_ = v.BindPFlag(dotReadHeaderTimeout, f.Lookup(dashReadHeaderTimeout))

	// _ = v.BindPFlag(dotHost, f.Lookup(dashHost))
	// _ = v.BindPFlag(dotPort, f.Lookup(dashPort))
}

func ternary[T any](condition bool, ifOutput T, elseOutput T) T {
	if condition {
		return ifOutput
	}
	return elseOutput
}

func buildKey(prefix, key string) (string, string) {
	dash := ternary(len(prefix) == 0, key, prefix+"-"+kebabcase.Kebabcase(key))
	dot := ternary(len(prefix) == 0, key, prefix+"."+key)
	return dash, dot
}
