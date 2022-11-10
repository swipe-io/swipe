package metric

import (
	"strconv"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Interface struct {
	TypeName string
	LcName   string
	UcName   string
	ModuleID string
	Methods  []Method
}

type Label struct {
	Name string
	Key  string
}

type Method struct {
	Name      option.String
	Sig       *option.SignType
	IsEnabled bool
}

type Metric struct {
	w          writer.GoWriter
	importer   swipe.Importer
	interfaces []Interface
	isGateway  bool
	labels     []Label
}

func (m *Metric) SetLabels(labels []Label) *Metric {
	m.labels = labels
	return m
}

func (m *Metric) SetIsGateway(isGateway bool) *Metric {
	m.isGateway = isGateway
	return m
}

func (m *Metric) SetInterfaces(interfaces []Interface) *Metric {
	m.interfaces = interfaces
	return m
}

func (m *Metric) Build() []byte {
	if len(m.interfaces) == 0 {
		return nil
	}

	metricsPkg := m.importer.Import("metrics", "github.com/go-kit/kit/metrics")
	timePkg := m.importer.Import("time", "time")
	stdPrometheusPkg := m.importer.Import("prometheus", "github.com/prometheus/client_golang/prometheus")
	kitPrometheusPkg := m.importer.Import("prometheus", "github.com/go-kit/kit/metrics/prometheus")

	m.w.W("type metricOpts struct {\n")
	m.w.W("requestCount %s.Counter\n", metricsPkg)
	m.w.W("requestLatency %s.Histogram\n", metricsPkg)
	m.w.W("}\n\n")

	m.w.W("type MetricOption func(*metricOpts)\n\n")

	m.w.W("func MetricRequestLatency(requestLatency %s.Histogram) MetricOption {\nreturn func(o *metricOpts) {\no.requestLatency = requestLatency\n}\n}\n\n", metricsPkg)
	m.w.W("func MetricRequestCount(requestCount %s.Counter) MetricOption {\nreturn func(o *metricOpts) {\no.requestCount = requestCount\n}\n}\n\n", metricsPkg)

	for _, iface := range m.interfaces {
		ifaceName := iface.UcName
		if m.isGateway {
			ifaceName = strcase.ToCamel(iface.ModuleID) + ifaceName
		}
		ifaceTypeName := ifaceName + "Interface"
		middlewareFuncName := "Metric" + ifaceName + "Middleware"
		middlewareTypeName := ifaceName + "Middleware"
		middlewareNameType := ifaceName + "MetricMiddleware"

		m.w.W("type %s struct {\n", middlewareNameType)
		m.w.W("next %s\n", ifaceTypeName)
		m.w.W("opts *metricOpts\n")
		m.w.W("}\n\n")

		for _, method := range iface.Methods {
			m.w.W("func (s *%s) %s %s {\n", middlewareNameType, method.Name.Value, swipe.TypeString(method.Sig, false, m.importer))

			if method.IsEnabled {

				methodName := method.Name.Value
				m.w.WriteDefer(
					[]string{"begin " + timePkg + ".Time"},
					[]string{timePkg + ".Now()"},
					func() {

						m.w.W("requestCount := s.opts.requestCount.With(\"method\", \"%s\")\n", methodName)
						m.w.W("requestLatency := s.opts.requestLatency.With(\"method\", \"%s\")\n", methodName)

						if len(m.labels) > 0 {
							m.w.W("requestCount = requestCount.With(")
							for i, l := range m.labels {
								if i > 0 {
									m.w.W(", ")
								}
								m.w.W("%s, ctx.Value(%s).(string)", strconv.Quote(l.Name), swipe.TypeString(l.Key, false, m.importer))
							}
							m.w.W(")\n")
							m.w.W("requestLatency = requestLatency.With(")
							for i, l := range m.labels {
								if i > 0 {
									m.w.W(", ")
								}
								m.w.W("%s, ctx.Value(%s).(string)", strconv.Quote(l.Name), swipe.TypeString(l.Key, false, m.importer))
							}
							m.w.W(")\n")
						}
						e := plugin.Error(method.Sig.Results)
						if e != nil {
							m.w.W("if %[1]s != nil {\nrequestCount.With(\"err\", %[1]s.Error()).Add(1)\n} else {\n", e.Name)
							m.w.W("")
						}
						m.w.W("requestCount.With(\"err\", \"\").Add(1)\n")
						if e != nil {
							m.w.W("}\n")
						}
						m.w.W("requestLatency.Observe(%s.Since(begin).Seconds())\n", timePkg)
					},
				)
			}
			if len(method.Sig.Results) > 0 {
				for i, result := range method.Sig.Results {
					if i > 0 {
						m.w.W(",")
					}
					m.w.W(result.Name.Value)
				}
				m.w.W(" = ")
			}

			m.w.W("s.next.%s(", method.Name)
			for i, param := range method.Sig.Params {
				if i > 0 {
					m.w.W(",")
				}
				var variadic string
				if param.IsVariadic {
					variadic = "..."
				}
				m.w.W(param.Name.Value + variadic)
			}
			m.w.W(")\n")

			m.w.W("return\n")

			m.w.W("}\n\n")
		}

		m.w.W("func %[1]s(namespace, subsystem string, opts ...MetricOption) %[3]s {\nreturn func(next %[2]s) %[2]s {\n", middlewareFuncName, ifaceTypeName, middlewareTypeName)

		m.w.W("i := &%s{next: next, opts: &metricOpts{}}\n", middlewareNameType)

		m.w.W("for _, o := range opts {\no(i.opts)\n}\n")

		m.w.W("if i.opts.requestCount == nil {\n")
		m.w.W("i.opts.requestCount = %s.NewCounterFrom(%s.CounterOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
		m.w.W("Namespace: namespace,\n")
		m.w.W("Subsystem: subsystem,\n")
		m.w.W("Name: %s,\n", strconv.Quote("request_count"))
		m.w.W("Help: %s,\n", strconv.Quote("Number of requests received."))

		m.w.W("}, []string{\"method\", \"err\"")

		for _, l := range m.labels {
			m.w.W(", %s", strconv.Quote(l.Name))
		}

		m.w.W("})\n")

		m.w.W("\n}\n")

		m.w.W("if i.opts.requestLatency == nil {\n")
		m.w.W("i.opts.requestLatency = %s.NewSummaryFrom(%s.SummaryOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
		m.w.W("Namespace: namespace,\n")
		m.w.W("Subsystem: subsystem,\n")
		m.w.W("Name: %s,\n", strconv.Quote("request_latency_microseconds"))
		m.w.W("Help: %s,\n", strconv.Quote("Total duration of requests in microseconds."))
		m.w.W("}, []string{\"method\"")

		for _, l := range m.labels {
			m.w.W(", %s", strconv.Quote(l.Name))
		}

		m.w.W("})\n")
		m.w.W("\n}\n")
		m.w.W("return i\n}\n}\n")
	}
	return m.w.Bytes()
}

func NewMetric(importer swipe.Importer) *Metric {
	return &Metric{importer: importer}
}
