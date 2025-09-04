package httpclient

import (
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// config contains configuration options for a http.Client.
type config struct {
	transport http.RoundTripper
	logger    *slog.Logger

	timeout           time.Duration
	keepAliveTimeout  time.Duration
	keepAliveInterval time.Duration
	keepAliveMax      int
	proxyURL          *url.URL
	tlsCert           []byte
	tlsKey            []byte
	userAgent         string
	insecure          bool

	disableOpenTelemetry bool
}

// newConfig returns a config configured with options.
func newConfig(options []option) *config {
	conf := &config{
		timeout:           30 * time.Second,
		keepAliveTimeout:  30 * time.Second,
		keepAliveInterval: 15 * time.Second,
		keepAliveMax:      256,
	}

	for _, o := range options {
		o.apply(conf)
	}

	return conf
}

type option interface{ apply(*config) }
type optionFunc func(*config)

func (o optionFunc) apply(conf *config) { o(conf) }

func WithTimeout(timeout time.Duration) option {
	return optionFunc(func(conf *config) {
		conf.timeout = timeout
	})
}

func WithKeepAliveTimeout(duration time.Duration) option {
	return optionFunc(func(conf *config) {
		conf.keepAliveTimeout = duration
	})
}

func WithKeepAliveInterval(interval time.Duration) option {
	return optionFunc(func(conf *config) {
		conf.keepAliveInterval = interval
	})
}

func WithKeepAliveMax(max int) option {
	return optionFunc(func(conf *config) {
		conf.keepAliveMax = max
	})
}

func WithTLSCertificate(cert, key []byte) option {
	return optionFunc(func(conf *config) {
		conf.tlsCert = make([]byte, len(cert))
		conf.tlsKey = make([]byte, len(key))
		copy(conf.tlsCert, cert)
		copy(conf.tlsKey, key)
	})
}

func WithInsecure(insecure ...bool) option {
	return optionFunc(func(conf *config) {
		if len(insecure) > 0 {
			conf.insecure = insecure[0]
		} else {
			conf.insecure = true
		}
	})
}

func WithProxyURL(u *url.URL) option {
	return optionFunc(func(conf *config) {
		conf.proxyURL = u
	})
}

func WithDisableOpenTelemetry() option {
	return optionFunc(func(conf *config) {
		conf.disableOpenTelemetry = true
	})
}

func WithLogger(l *slog.Logger) option {
	return optionFunc(func(conf *config) {
		conf.logger = l
	})
}

func WithUserAgent(ua string) option {
	return optionFunc(func(conf *config) {
		conf.userAgent = ua
	})
}
