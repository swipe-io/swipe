//+build !swipe

// Code generated by Swipe v2.0.0-rc4. DO NOT EDIT.

package app

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	prometheus2 "github.com/go-kit/kit/metrics/prometheus"
	"github.com/prometheus/client_golang/prometheus"
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

type AppInterfaceInstrumentingMiddleware struct {
	next AppInterface
	opts *instrumentingOpts
}

func (s *AppInterfaceInstrumentingMiddleware) Create(ctx context.Context, newData Data, name string, data []byte, date time.Time) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Create").Add(1)
		s.opts.requestLatency.With("method", "Create").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Create(ctx, newData, name, data, date)
}

func (s *AppInterfaceInstrumentingMiddleware) Delete(ctx context.Context, id uint) (string, string, error) {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Delete").Add(1)
		s.opts.requestLatency.With("method", "Delete").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Delete(ctx, id)
}

func (s *AppInterfaceInstrumentingMiddleware) Get(ctx context.Context, id int, name string, fname string, price float32, n int, b int, cc int) (User, error) {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Get").Add(1)
		s.opts.requestLatency.With("method", "Get").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Get(ctx, id, name, fname, price, n, b, cc)
}

func (s *AppInterfaceInstrumentingMiddleware) GetAll(ctx context.Context, members Members) ([]*User, error) {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "GetAll").Add(1)
		s.opts.requestLatency.With("method", "GetAll").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.GetAll(ctx, members)
}

func (s *AppInterfaceInstrumentingMiddleware) Start(ctx context.Context) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "Start").Add(1)
		s.opts.requestLatency.With("method", "Start").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.Start(ctx)
}

func (s *AppInterfaceInstrumentingMiddleware) TestMethod(data map[string]interface{}, ss interface{}) (map[string]map[int][]string, error) {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "TestMethod").Add(1)
		s.opts.requestLatency.With("method", "TestMethod").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.TestMethod(data, ss)
}

func (s *AppInterfaceInstrumentingMiddleware) TestMethod2(ctx context.Context, ns string, utype string, user string, restype string, resource string, permission string) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "TestMethod2").Add(1)
		s.opts.requestLatency.With("method", "TestMethod2").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.TestMethod2(ctx, ns, utype, user, restype, resource, permission)
}

func (s *AppInterfaceInstrumentingMiddleware) TestMethodOptionals(ctx context.Context, ns string) error {
	defer func(begin time.Time) {
		s.opts.requestCount.With("method", "TestMethodOptionals").Add(1)
		s.opts.requestLatency.With("method", "TestMethodOptionals").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.next.TestMethodOptionals(ctx, ns)
}

func NewInstrumentingAppInterfaceMiddleware(s AppInterface, opts ...InstrumentingOption) AppInterface {
	i := &AppInterfaceInstrumentingMiddleware{next: s, opts: &instrumentingOpts{}}
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