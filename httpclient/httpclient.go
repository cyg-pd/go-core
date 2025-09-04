package httpclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/samber/lo"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func New(opts ...option) *http.Client {
	conf := newConfig(opts)
	client := &http.Client{
		Transport: conf.transport,
		Timeout:   conf.timeout,
	}

	if client.Transport == nil {
		client.Transport = newTransport(conf)
	}

	return client
}

func NewTransport(opts ...option) (t http.RoundTripper) {
	conf := newConfig(opts)
	return newTransport(conf)
}

func newProxy(conf *config) func(*http.Request) (*url.URL, error) {
	if conf.proxyURL != nil {
		return http.ProxyURL(conf.proxyURL)
	}
	return http.ProxyFromEnvironment
}

func newTransport(conf *config) (t http.RoundTripper) {
	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: conf.insecure, //nolint:gosec
	}

	if len(conf.tlsCert) != 0 && len(conf.tlsKey) != 0 {
		tlsConf.Certificates = []tls.Certificate{newClientCert(conf)}
	}

	t = &http.Transport{
		Proxy:                 newProxy(conf),
		DialContext:           newDialerContext(conf),
		MaxIdleConnsPerHost:   conf.keepAliveMax,
		MaxIdleConns:          conf.keepAliveMax,
		IdleConnTimeout:       conf.keepAliveTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       tlsConf,
	}

	if len(conf.userAgent) > 0 {
		t = &appendUserAgentTransport{
			Transport: t,
			AppendUA:  conf.userAgent,
		}
	}

	if conf.logger != nil {
		t = newRequestLogTransport(t, conf.logger)
	}

	if !conf.disableOpenTelemetry {
		t = otelhttp.NewTransport(t)
	}

	return t
}

func newDialerContext(conf *config) func(context.Context, string, string) (net.Conn, error) {
	dialContext := reportMetric(newDialer(conf).DialContext)
	return dialContext
}

func newDialer(conf *config) *net.Dialer {
	return &net.Dialer{
		Timeout:   conf.timeout,
		KeepAlive: conf.keepAliveInterval,
	}
}

func newClientCert(conf *config) tls.Certificate {
	return lo.Must(tls.X509KeyPair(conf.tlsCert, conf.tlsKey))
}
