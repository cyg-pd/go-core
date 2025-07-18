package httpclient

import (
	"context"
	"net"
	"sync"
	"sync/atomic"

	"github.com/cyg-pd/go-otelx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var openConnCount sync.Map
var meter = otelx.Meter()

func init() {
	openConns, _ := meter.Int64ObservableGauge(
		"http.client.open_connections",
		metric.WithDescription("The number of established connections both in use and idle"),
	)

	if _, err := meter.RegisterCallback(
		func(ctx context.Context, o metric.Observer) error {
			openConnCount.Range(func(key, value any) bool {
				if addr, port, err := net.SplitHostPort(key.(string)); err == nil {
					o.ObserveInt64(
						openConns,
						value.(*atomic.Int64).Load(),
						metric.WithAttributes(
							attribute.String("net.peer.name", addr),
							attribute.String("net.peer.port", port),
						),
					)
				}
				return true
			})
			return nil
		},
		openConns,
	); err != nil {
		panic(err)
	}
}

type connCounter struct {
	net.Conn
	network, address string
	closeCounter     sync.Once
	counter          *atomic.Int64
}

func (conn *connCounter) Close() error {
	defer func() {
		conn.closeCounter.Do(func() {
			conn.counter.Add(-1)
		})
	}()
	return conn.Conn.Close()
}

func reportMetric(fn func(context.Context, string, string) (net.Conn, error)) func(context.Context, string, string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		conn, err := fn(ctx, network, address)
		if err == nil && conn != nil {
			v, ok := openConnCount.Load(address)
			if !ok {
				v, _ = openConnCount.LoadOrStore(address, &atomic.Int64{})
			}
			v.(*atomic.Int64).Add(1)
			return &connCounter{
				Conn:    conn,
				network: network,
				address: address,
				counter: v.(*atomic.Int64),
			}, nil
		}
		return conn, err
	}
}
