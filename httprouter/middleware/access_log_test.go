package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func mockSlog(buf *bytes.Buffer) (fn func()) {
	fn = func(s *slog.Logger) func() { return func() { slog.SetDefault(s) } }(slog.Default())
	slog.SetDefault(slog.New(slog.NewJSONHandler(buf, nil)))
	return
}

var logTemplate = `
{
	"package": "` + pkg + `",
	"channel": "access",
	%s
}`

type brokenReader struct{ err error }

func (br *brokenReader) Read(p []byte) (n int, err error) {
	return 0, br.err
}

type header struct {
	Key   string
	Value string
}

func createRouter() (router *gin.Engine) {
	gin.SetMode(gin.TestMode)
	router = gin.New()
	router.Use(NewAccessLog().Middleware())
	router.POST("/access", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{"msg": "test"})
	})
	router.POST("/error", func(ctx *gin.Context) {
		ctx.AbortWithError(500, errors.New("test error"))
	})
	return
}

func PerformRequest(r http.Handler, method, path string, body io.Reader, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, body)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAccessLogSkip(t *testing.T) {
	buf := &bytes.Buffer{}
	defer mockSlog(buf)()

	router := gin.New()
	router.Use(NewAccessLog(WithAccessLogFilter(func(r *http.Request) bool {
		return !strings.HasPrefix(r.URL.Path, "/access")
	})).Middleware())
	router.POST("/access", func(ctx *gin.Context) {
		ctx.JSON(200, map[string]string{})
	})

	// RUN
	PerformRequest(router, http.MethodPost, "https://example.com/access", nil)

	// TEST
	is := assert.New(t)
	var logData map[string]any
	json.Unmarshal(buf.Bytes(), &logData)
	is.Equal("", buf.String())
}

func TestNoReqBody(t *testing.T) {
	buf := &bytes.Buffer{}
	defer mockSlog(buf)()

	router := createRouter()

	// RUN
	PerformRequest(router, http.MethodPost, "https://example.com/access?q=1&qq=2", nil, header{"k1", "v1"})

	// TEST
	is := assert.New(t)
	var logData map[string]any
	json.Unmarshal(buf.Bytes(), &logData)

	is.JSONEq(fmt.Sprintf(logTemplate, `
	"ip": "`+cast.ToString(logData["ip"])+`",
	"latency": `+cast.ToString(logData["latency"])+`,
	"level": "INFO",
	"msg": "POST /access HTTP/1.1",
	"req": {
		"method": "POST",
		"host": "example.com",
		"path": "/access",
		"url": "https://example.com/access?q=1&qq=2",
		"headers": {
			"K1": [
				"v1"
			]
		},
		"query": {
			"q": [
				"1"
			],
			"qq": [
				"2"
			]
		}
	},
	"res": {
		"status_code": 200,
		"headers": {
			"Content-Type": ["application/json; charset=utf-8"]
		},
		"body": "{\"msg\":\"test\"}",
		"size": 14
	},
	"route": "/access",
	"time": "`+cast.ToString(logData["time"])+`"
`), buf.String())
}

func TestInputReqBody(t *testing.T) {
	buf := &bytes.Buffer{}
	defer mockSlog(buf)()

	router := createRouter()

	body := strings.NewReader(`{"title":"Access log Middleware","content":"Test body"}`)
	// RUN
	PerformRequest(router, http.MethodPost, "https://example.com/access?q=1&qq=2", body, header{"k1", "v1"})

	// TEST
	is := assert.New(t)
	var logData map[string]any
	json.Unmarshal(buf.Bytes(), &logData)

	is.JSONEq(fmt.Sprintf(logTemplate, `
	"ip": "`+cast.ToString(logData["ip"])+`",
	"latency": `+cast.ToString(logData["latency"])+`,
	"level": "INFO",
	"msg": "POST /access HTTP/1.1",
	"req": {
		"method": "POST",
		"host": "example.com",
		"path": "/access",
		"url": "https://example.com/access?q=1&qq=2",
		"body": "{\"title\":\"Access log Middleware\",\"content\":\"Test body\"}",
		"headers": {
			"K1": [
				"v1"
			]
		},
		"query": {
			"q": [
				"1"
			],
			"qq": [
				"2"
			]
		}
	},
	"res": {
		"status_code": 200,
		"headers": {
			"Content-Type": ["application/json; charset=utf-8"]
		},
		"body": "{\"msg\":\"test\"}",
		"size": 14
	},
	"route": "/access",
	"time": "`+cast.ToString(logData["time"])+`"
`), buf.String())
}

func TestReqBodyReadError(t *testing.T) {
	buf := &bytes.Buffer{}
	defer mockSlog(buf)()

	err := errors.New("failed reading")
	body := io.NopCloser(&brokenReader{err})
	router := createRouter()

	// RUN
	PerformRequest(router, http.MethodPost, "https://example.com/readError", body)

	// TEST
	is := assert.New(t)
	var logData map[string]any
	json.Unmarshal(buf.Bytes(), &logData)

	is.JSONEq(fmt.Sprintf(logTemplate, `
	"ip": "`+cast.ToString(logData["ip"])+`",
	"latency": `+cast.ToString(logData["latency"])+`,
	"level": "INFO",
	"msg": "POST /readError HTTP/1.1",
	"req": {
		"method": "POST",
		"host": "example.com",
		"path": "/readError",
		"url": "https://example.com/readError",
		"headers": {},
		"query": {}
	},
	"res": {
		"status_code": 404,
		"headers": {},
		"body": "",
		"size": 0
	},
	"route": "",
	"time": "`+cast.ToString(logData["time"])+`"
`), buf.String())
}

func TestGinResError(t *testing.T) {
	buf := &bytes.Buffer{}
	defer mockSlog(buf)()

	router := createRouter()

	// RUN
	PerformRequest(router, http.MethodPost, "https://example.com/error", nil)

	// TEST
	is := assert.New(t)
	var logData map[string]any
	json.Unmarshal(buf.Bytes(), &logData)

	is.JSONEq(fmt.Sprintf(logTemplate, `
	"go.error": "Error #01: test error\n",
	"ip": "`+cast.ToString(logData["ip"])+`",
	"latency": `+cast.ToString(logData["latency"])+`,
	"level": "ERROR",
	"msg": "POST /error HTTP/1.1",
	"req": {
		"method": "POST",
		"host": "example.com",
		"path": "/error",
		"url": "https://example.com/error",
		"headers": {},
		"query": {}
	},
	"res": {
		"status_code": 500,
		"headers": {},
		"body": "",
		"size": 0
	},
	"route": "/error",
	"time": "`+cast.ToString(logData["time"])+`"
`), buf.String())
}
