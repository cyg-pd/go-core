package httprouter

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetupFlags(f *pflag.FlagSet, v *viper.Viper, prefix string) {
	dashHost := ternary(len(prefix) == 0, "host", prefix+"-host")
	dashPort := ternary(len(prefix) == 0, "port", prefix+"-port")
	dotHost := ternary(len(prefix) == 0, "host", prefix+".host")
	dotPort := ternary(len(prefix) == 0, "port", prefix+".port")

	f.String(dashHost, "127.0.0.1", "HTTP Server listen host")
	f.Int(dashPort, 8080, "HTTP Server listen port")

	_ = v.BindPFlag(dotHost, f.Lookup(dashHost))
	_ = v.BindPFlag(dotPort, f.Lookup(dashPort))
}

func ternary[T any](condition bool, ifOutput T, elseOutput T) T {
	if condition {
		return ifOutput
	}
	return elseOutput
}
