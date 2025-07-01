package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// accessLogOption applies a configuration accessLogOption value to a http.Client.
type accessLogOption interface {
	apply(*AccessLog)
}

// accessLogOptionFunc applies a set of options to a config.
type accessLogOptionFunc func(*AccessLog)

// apply returns a config with option(s) applied.
func (o accessLogOptionFunc) apply(conf *AccessLog) {
	o(conf)
}

// WithAccessLogFilter associates slog.Logger with a http.Client.
func WithAccessLogFilter(filters ...func(r *http.Request) bool) accessLogOption {
	return accessLogOptionFunc(func(conf *AccessLog) {
		conf.filters = filters
	})
}

func NewAccessLog(opts ...accessLogOption) *AccessLog {
	m := &AccessLog{}
	for _, opt := range opts {
		opt.apply(m)
	}
	return m
}

type AccessLog struct {
	filters []func(r *http.Request) bool
}

func (a AccessLog) reqBody(req *http.Request) slog.Attr {
	if req.Body == nil || req.Body == http.NoBody {
		return slog.Attr{}
	}

	// content length bigger then 100kb ignore read body
	if req.ContentLength > 100*1024 {
		return slog.Attr{}
	}

	b, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return slog.Attr{}
	}
	req.Body = io.NopCloser(bytes.NewReader(b))

	return slog.String("body", string(b))
}

func (a *AccessLog) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, f := range a.filters {
			if !f(c.Request) {
				// Serve the request to the next middleware
				// if a filter rejects the request.
				c.Next()
				return
			}
		}

		// Start timer
		start := time.Now()

		// req body
		reqBody := a.reqBody(c.Request)

		// res body
		buf := &bytes.Buffer{}
		// buf := buffpool.Get()
		buf.Reset()
		defer func(w gin.ResponseWriter) {
			c.Writer = w
			// buffpool.Put(buf)
		}(c.Writer)
		c.Writer = &bodyLogWriter{buf: buf, ResponseWriter: c.Writer}

		// Process request
		c.Next()
		if err := c.Errors.String(); err == "" {
			a.log(slog.InfoContext, c, time.Since(start), reqBody, buf, slog.Attr{})
		} else {
			a.log(slog.ErrorContext, c, time.Since(start), reqBody, buf, slog.String("go.error", err))
		}
	}
}

func (a *AccessLog) log(
	log func(context.Context, string, ...any),
	c *gin.Context,
	latency time.Duration,
	reqBody slog.Attr,
	buf *bytes.Buffer,
	err slog.Attr,
) {

	log(
		c.Request.Context(),
		c.Request.Method+" "+c.Request.URL.Path+" "+c.Request.Proto,
		err,
		slog.String("package", pkg),
		slog.String("channel", "access"),
		slog.String("ip", c.ClientIP()),
		slog.Int64("latency", latency.Milliseconds()),
		slog.String("route", c.FullPath()),

		// Request Group
		slog.Group("req",
			slog.String("method", c.Request.Method),
			slog.String("url", c.Request.URL.String()),
			slog.String("host", c.Request.Host),
			slog.String("path", c.Request.URL.Path),
			slog.Any("headers", c.Request.Header),
			reqBody,
			slog.Any("query", c.Request.URL.Query()),
		),

		// Response Group
		slog.Group("res",
			slog.Int("status_code", c.Writer.Status()),
			slog.Any("headers", c.Writer.Header()),
			slog.String("body", buf.String()),
			slog.Int("size", len(buf.Bytes())),
		),
	)
}

type bodyLogWriter struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}
