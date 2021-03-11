//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	prometheus2 "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/swipe-io/swipe/v2/fixture/ServiceRESTMultiIdenticalInterface/app/controller/app1"
	"github.com/swipe-io/swipe/v2/fixture/ServiceRESTMultiIdenticalInterface/app/controller/app2"
)

type instrumentingOpts struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	namespace      string
	subsystem      string
}

type InstrumentingOption func(*instrumentingOpts)

func Namespace(v string) InstrumentingOption {
	return func(o *instrumentingOpts) {
		o.namespace = v
	}
}

func Subsystem(v string) InstrumentingOption {
	return func(o *instrumentingOpts) {
		o.subsystem = v
	}
}

func RequestLatency(requestLatency metrics.Histogram) InstrumentingOption {
	return func(o *instrumentingOpts) {
		o.requestLatency = requestLatency
	}
}

func RequestCount(requestCount metrics.Counter) InstrumentingOption {
	return func(o *instrumentingOpts) {
		o.requestCount = requestCount
	}
}

type App1InstrumentingMiddleware struct {
	next app1.App
	opts *instrumentingOpts
}

func (s *App1InstrumentingMiddleware) Create(ctx context.Context, name string, data []byte) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Create").Add(1)
		s.opts.requestLatency.With("method", "Create").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Create(ctx, name, data)
}

func NewInstrumentingApp1Middleware(s app1.App, opts ...InstrumentingOption) app1.App {
	i := &App1InstrumentingMiddleware{next: s, opts: &instrumentingOpts{}}
	for _, o := range opts {
		o(i.opts)
	}
	if i.opts.requestCount == nil {
		i.opts.requestCount = prometheus2.NewCounterFrom(prometheus.CounterOpts{
			Namespace: i.opts.namespace,
			Subsystem: i.opts.subsystem,
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"})

	}
	if i.opts.requestLatency == nil {
		i.opts.requestLatency = prometheus2.NewSummaryFrom(prometheus.SummaryOpts{
			Namespace: i.opts.namespace,
			Subsystem: i.opts.subsystem,
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"})

	}
	return i
}

type App2InstrumentingMiddleware struct {
	next app2.App
	opts *instrumentingOpts
}

func (s *App2InstrumentingMiddleware) Create(ctx context.Context, name string, data []byte) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Create").Add(1)
		s.opts.requestLatency.With("method", "Create").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Create(ctx, name, data)
}

func NewInstrumentingApp2Middleware(s app2.App, opts ...InstrumentingOption) app2.App {
	i := &App2InstrumentingMiddleware{next: s, opts: &instrumentingOpts{}}
	for _, o := range opts {
		o(i.opts)
	}
	if i.opts.requestCount == nil {
		i.opts.requestCount = prometheus2.NewCounterFrom(prometheus.CounterOpts{
			Namespace: i.opts.namespace,
			Subsystem: i.opts.subsystem,
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"})

	}
	if i.opts.requestLatency == nil {
		i.opts.requestLatency = prometheus2.NewSummaryFrom(prometheus.SummaryOpts{
			Namespace: i.opts.namespace,
			Subsystem: i.opts.subsystem,
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"})

	}
	return i
}