package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

const logTemplate = `
{
	"channel": "request",
	"time": "%s",
	"latency": %s,
	%s
}
`

type mockRoundTripper struct {
	res *http.Response
	e   error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (res *http.Response, e error) {
	return m.res, m.e
}

type brokenReader struct{ err error }

func (br *brokenReader) Read(p []byte) (n int, err error) {
	return 0, br.err
}

func TestSuccess(t *testing.T) {
	buf := &bytes.Buffer{}
	body := `{"Currency":"","CashBet":0.00}`

	r := newRequestLogTransport(
		&mockRoundTripper{res: &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBufferString(body))}},
		&config{
			logger:    slog.New(slog.NewJSONHandler(buf, nil)),
			logOption: []logOption{WithMaxBodySize(0)},
		},
	)

	req, _ := http.NewRequest(
		http.MethodPost,
		"https://example.com/platform/feed/summary?currency=CNY&interval=0",
		bytes.NewBufferString(`{"key": "value"}`),
	)

	res, e := r.RoundTrip(req) //nolint:bodyclose
	is := assert.New(t)
	is.Nil(e)

	resBody, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()

	is.JSONEq(body, string(resBody))

	var log map[string]any
	is.NoError(json.Unmarshal(buf.Bytes(), &log))

	is.JSONEq(
		fmt.Sprintf(logTemplate, cast.ToString(log["time"]), cast.ToString(log["latency"]), `
			"level": "INFO",
			"msg": "POST /platform/feed/summary HTTP/1.1",
			"req": {
				"method": "POST",
				"host": "example.com",
				"url": "https://example.com/platform/feed/summary?currency=CNY&interval=0",
				"body": "{\"key\": \"value\"}",
				"path": "/platform/feed/summary",
				"query": {
					"currency": [
						"CNY"
					],
					"interval": [
						"0"
					]
				},
				"headers": {}
			},
			"res": {
				"status_code": 200,
				"headers":null,
				"size": 30,
				"body": "{\"Currency\":\"\",\"CashBet\":0.00}"
			}
		`),
		buf.String(),
	)
}

func TestBodyReadError(t *testing.T) {
	buf := &bytes.Buffer{}
	err := errors.New("failed reading")

	body := io.NopCloser(&brokenReader{err})

	r := newRequestLogTransport(
		&mockRoundTripper{res: &http.Response{StatusCode: 200, Body: body}},
		&config{
			logger: slog.New(slog.NewJSONHandler(buf, nil)),
		},
	)

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/platform/feed/summary?currency=CNY&interval=0", nil)

	_, e := r.RoundTrip(req) //nolint:bodyclose
	is := assert.New(t)
	is.NotNil(e)
	is.ErrorIs(e, err)

	var log map[string]any
	is.NoError(json.Unmarshal(buf.Bytes(), &log))

	is.JSONEq(
		fmt.Sprintf(logTemplate, cast.ToString(log["time"]), cast.ToString(log["latency"]), `
			"level": "ERROR",
			"msg": "GET /platform/feed/summary HTTP/1.1",
			"req": {
				"method": "GET",
				"host": "example.com",
				"url": "https://example.com/platform/feed/summary?currency=CNY&interval=0",
				"path": "/platform/feed/summary",
				"query": {
					"currency": [
						"CNY"
					],
					"interval": [
						"0"
					]
				},
				"headers": {}
			},
			"res": {
				"status_code": 200,
				"headers": null
			},
			"error.message": "read response body fail: failed reading"
		`),
		buf.String(),
	)
}

func TestRoundTripError(t *testing.T) {
	buf := &bytes.Buffer{}

	err := errors.New("error")
	r := newRequestLogTransport(
		&mockRoundTripper{e: err},
		&config{
			logger: slog.New(slog.NewJSONHandler(buf, nil)),
		},
	)

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/platform/feed/summary?currency=CNY&interval=0", nil)

	_, e := r.RoundTrip(req) //nolint:bodyclose
	is := assert.New(t)
	is.ErrorIs(e, err)

	var log map[string]any
	is.NoError(json.Unmarshal(buf.Bytes(), &log))

	is.JSONEq(
		fmt.Sprintf(logTemplate, cast.ToString(log["time"]), cast.ToString(log["latency"]), `
			"level": "ERROR",
			"msg": "GET /platform/feed/summary HTTP/1.1",
			"req": {
				"method": "GET",
				"host": "example.com",
				"url": "https://example.com/platform/feed/summary?currency=CNY&interval=0",
				"path": "/platform/feed/summary",
				"query": {
					"currency": [
						"CNY"
					],
					"interval": [
						"0"
					]
				},
				"headers": {}
			},
			"error.message": "error"
		`),
		buf.String(),
	)
}
