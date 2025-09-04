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
type accessLogOption interface{ apply(*AccessLog) }

// accessLogOptionFunc applies a set of options to a config.
type accessLogOptionFunc func(*AccessLog)

// apply returns a config with option(s) applied.
func (o accessLogOptionFunc) apply(conf *AccessLog) { o(conf) }

// WithAccessLogFilter associates slog.Logger with a http.Client.
func WithAccessLogFilter(filters ...func(r *http.Request) bool) accessLogOption {
	return accessLogOptionFunc(func(conf *AccessLog) {
		conf.filters = filters
	})
}

func WithAccessLogMaxBodySize(maxSize int) accessLogOption {
	return accessLogOptionFunc(func(conf *AccessLog) {
		conf.maxBodySize = int64(maxSize)
	})
}

func NewAccessLog(opts ...accessLogOption) *AccessLog {
	m := &AccessLog{
		maxBodySize: -1,
	}
	for _, opt := range opts {
		opt.apply(m)
	}
	return m
}

type AccessLog struct {
	maxBodySize int64
	filters     []func(r *http.Request) bool
}

func (a AccessLog) reqBody(req *http.Request) slog.Attr {
	if a.maxBodySize < 0 {
		return slog.Attr{}
	}

	if req.Body == nil || req.Body == http.NoBody {
		return slog.Attr{}
	}

	// content length bigger then 100kb ignore read body
	if a.maxBodySize > 0 && req.ContentLength > a.maxBodySize {
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

func (a AccessLog) resBody(c *gin.Context) slog.Attr {
	if a.maxBodySize < 0 {
		c.Next()
		return slog.Attr{}
	}

	buf := &bytes.Buffer{}
	buf.Reset()

	defer func(w gin.ResponseWriter) { c.Writer = w }(c.Writer)
	c.Writer = &bodyLogWriter{buf: buf, ResponseWriter: c.Writer}
	c.Next()

	if a.maxBodySize > 0 && int64(buf.Len()) > a.maxBodySize {
		return slog.Attr{}
	}

	return slog.String("body", buf.String())
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

		reqBody := a.reqBody(c.Request)
		start := time.Now()
		resBody := a.resBody(c)
		finish := time.Since(start)

		status := c.Writer.Status()
		switch {
		case status >= 500:
			a.log(slog.LevelError, c, finish, reqBody, resBody)
		case status >= 400:
			a.log(slog.LevelWarn, c, finish, reqBody, resBody)
		default:
			a.log(slog.LevelInfo, c, finish, reqBody, resBody)
		}
	}
}

func (a *AccessLog) log(
	lvl slog.Level,
	c *gin.Context,
	latency time.Duration,
	reqBody slog.Attr,
	resBody slog.Attr,
) {

	err := slog.Attr{}
	if s := c.Errors.String(); len(s) > 0 {
		err = slog.String("error.message", s)
	}

	slog.Log(
		context.WithoutCancel(c.Request.Context()),
		lvl,
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
			resBody,
			slog.Int("size", c.Writer.Size()),
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
