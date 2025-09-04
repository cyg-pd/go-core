package httpclient

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func newRequestLogTransport(r http.RoundTripper, l *config) http.RoundTripper {
	if l.logger == nil {
		return r
	}

	rt := &requestLog{
		Proxied:     r,
		log:         l.logger,
		maxBodySize: -1,
	}

	for _, opt := range l.logOption {
		opt.apply(rt)
	}

	return rt
}

type requestLog struct {
	Proxied     http.RoundTripper
	log         *slog.Logger
	maxBodySize int
}

func (l requestLog) RoundTrip(req *http.Request) (res *http.Response, e error) {
	ctx := req.Context()
	reqBody := l.reqBody(req)

	start := time.Now()
	res, e = l.Proxied.RoundTrip(req)

	msg := req.Method + " " + req.URL.Path + " " + req.Proto
	ch := slog.String("channel", "request")
	lat := slog.Int64("latency", time.Since(start).Milliseconds())
	reqAttr := slog.Group("req",
		slog.String("method", req.Method),
		slog.String("host", req.Host),
		slog.String("url", req.URL.String()),
		slog.String("path", req.URL.Path),
		slog.Any("query", req.URL.Query()),
		slog.Any("headers", req.Header),
		reqBody,
	)

	if e != nil {
		l.log.ErrorContext(ctx, msg, ch, lat, reqAttr, slog.String("error.message", e.Error()))
		return res, e
	}

	resAttr, err := l.resAttr(res)
	if err != nil {
		l.log.ErrorContext(ctx, msg, ch, lat, reqAttr, resAttr, slog.String("error.message", err.Error()))
		return res, err
	}

	l.log.InfoContext(ctx, msg, ch, lat, reqAttr, resAttr)

	return res, e
}

func (l requestLog) reqBody(req *http.Request) slog.Attr {
	if l.maxBodySize < 0 {
		return slog.Attr{}
	}

	if req.Body == nil {
		return slog.Attr{}
	}

	rb, err := req.GetBody()
	if err != nil || rb == nil {
		return slog.Attr{}
	}

	b, err := io.ReadAll(rb)
	if err != nil || len(b) == 0 {
		return slog.Attr{}
	}

	if l.maxBodySize > 0 && len(b) > l.maxBodySize {
		return slog.Attr{}
	}

	return slog.String("body", string(b))
}

func (l requestLog) resBody(res *http.Response, raw []byte) slog.Attr {
	if l.maxBodySize < 0 {
		return slog.Attr{}
	}

	b := l.decodeBody(res, raw)
	if len(b) == 0 {
		return slog.Attr{}
	}

	if l.maxBodySize > 0 && len(b) > l.maxBodySize {
		return slog.Attr{}
	}

	return slog.String("body", string(b))
}

func (l requestLog) decodeBody(res *http.Response, raw []byte) []byte {
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		r, err1 := gzip.NewReader(io.NopCloser(bytes.NewBuffer(raw)))
		defer func() { _ = r.Close() }()
		if err1 != nil {
			return raw
		}
		b, err2 := io.ReadAll(r)
		if err2 != nil {
			return raw
		}
		return b
	default:
		return raw
	}
}

func (l requestLog) resAttr(res *http.Response) (slog.Attr, error) {
	resAttr := slog.Group("res",
		slog.Int("status_code", res.StatusCode),
		slog.Any("headers", res.Header),
	)

	if res.Body != nil {
		b, err := io.ReadAll(res.Body)
		_ = res.Body.Close()
		res.Body = io.NopCloser(bytes.NewBuffer(b))

		if len(b) > 0 {
			resAttr = slog.Group("res",
				l.resBody(res, b),
				slog.Int("size", len(b)),
				slog.Int("status_code", res.StatusCode),
				slog.Any("headers", res.Header),
			)
		}

		if err != nil {
			return resAttr, fmt.Errorf("read response body fail: %w", err)
		}
	}

	return resAttr, nil
}
