package httprouter

import (
	coreconfig "github.com/cyg-pd/go-core/config"
	"github.com/spf13/pflag"
)

func init() {
	pflag.String("http-host", "127.0.0.1", "HTTP Server listen host")
	pflag.Int("http-port", 8080, "HTTP Server listen port")

	coreconfig.BindPFlag("http.host", pflag.Lookup("http-host"))
	coreconfig.BindPFlag("http.port", pflag.Lookup("http-port"))
}
