package generator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/swipe-io/swipe/v3/internal/plugin"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Instrumenting struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
	Labels        []config.InstrumentingLabel
}

func (g *Instrumenting) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	metricsPkg := importer.Import("metrics", "github.com/go-kit/kit/metrics")

	g.w.W("type instrumentingOpts struct {\n")
	g.w.W("requestCount %s.Counter\n", metricsPkg)
	g.w.W("requestLatency %s.Histogram\n", metricsPkg)
	g.w.W("}\n\n")

	g.w.W("type InstrumentingOption func(*instrumentingOpts)\n\n")

	g.w.W("func InstrumentingRequestLatency(requestLatency %s.Histogram) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.requestLatency = requestLatency\n}\n}\n\n", metricsPkg)
	g.w.W("func InstrumentingRequestCount(requestCount %s.Counter) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.requestCount = requestCount\n}\n}\n\n", metricsPkg)

	if len(g.Interfaces) > 0 {
		timePkg := importer.Import("time", "time")
		stdPrometheusPkg := importer.Import("prometheus", "github.com/prometheus/client_golang/prometheus")
		kitPrometheusPkg := importer.Import("prometheus", "github.com/go-kit/kit/metrics/prometheus")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)

			ifaceTypeName := NameInterface(iface)
			middlewareFuncName := fmt.Sprintf("Instrumenting%sMiddleware", UcNameWithAppPrefix(iface))
			middlewareTypeName := IfaceMiddlewareTypeName(iface)
			middlewareNameType := NameInstrumentingMiddleware(iface)

			g.w.W("type %s struct {\n", middlewareNameType)
			g.w.W("next %s\n", ifaceTypeName)
			g.w.W("opts *instrumentingOpts\n")
			g.w.W("}\n\n")

			for _, m := range ifaceType.Methods {
				mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]
				g.w.W("func (s *%s) %s %s {\n", middlewareNameType, m.Name.Value, swipe.TypeString(m.Sig, false, importer))
				if mopt.Instrumenting.Take() {
					methodName := iface.Named.Name.Lower() + "." + m.Name.Value
					g.w.WriteDefer(
						[]string{"begin " + timePkg + ".Time"},
						[]string{timePkg + ".Now()"},
						func() {

							g.w.W("requestCount := s.opts.requestCount.With(\"method\", \"%s\")\n", methodName)
							g.w.W("requestLatency := s.opts.requestLatency.With(\"method\", \"%s\")\n", methodName)

							if len(g.Labels) > 0 {
								g.w.W("requestCount = requestCount.With(")
								for i, l := range g.Labels {
									if i > 0 {
										g.w.W(", ")
									}
									g.w.W("%s, ctx.Value(%s).(string)", strconv.Quote(l.Name), swipe.TypeString(l.Key, false, importer))
								}
								g.w.W(")\n")
								g.w.W("requestLatency = requestLatency.With(")
								for i, l := range g.Labels {
									if i > 0 {
										g.w.W(", ")
									}
									g.w.W("%s, ctx.Value(%s).(string)", strconv.Quote(l.Name), swipe.TypeString(l.Key, false, importer))
								}
								g.w.W(")\n")
							}
							e := plugin.Error(m.Sig.Results)
							if e != nil {
								g.w.W("if %[1]s != nil {\nrequestCount.With(\"err\", %[1]s.Error()).Add(1)\n} else {\n", e.Name)
								g.w.W("")
							}
							g.w.W("requestCount.With(\"err\", \"\").Add(1)\n")
							if e != nil {
								g.w.W("}\n")
							}
							g.w.W("requestLatency.Observe(%s.Since(begin).Seconds())\n", timePkg)
						},
					)
				}
				if len(m.Sig.Results) > 0 {
					for i, result := range m.Sig.Results {
						if i > 0 {
							g.w.W(",")
						}
						g.w.W(result.Name.Value)
					}
					g.w.W(" = ")
				}

				g.w.W("s.next.%s(", m.Name)
				for i, param := range m.Sig.Params {
					if i > 0 {
						g.w.W(",")
					}
					var variadic string
					if param.IsVariadic {
						variadic = "..."
					}
					g.w.W(param.Name.Value + variadic)
				}
				g.w.W(")\n")

				g.w.W("return\n")

				g.w.W("}\n")
			}

			g.w.W("func %[1]s(namespace, subsystem string, opts ...InstrumentingOption) %[3]s {\nreturn func(next %[2]s) %[2]s {\n", middlewareFuncName, ifaceTypeName, middlewareTypeName)

			g.w.W("i := &%s{next: next, opts: &instrumentingOpts{}}\n", middlewareNameType)

			g.w.W("for _, o := range opts {\no(i.opts)\n}\n")

			g.w.W("if i.opts.requestCount == nil {\n")
			g.w.W("i.opts.requestCount = %s.NewCounterFrom(%s.CounterOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
			g.w.W("Namespace: namespace,\n")
			g.w.W("Subsystem: subsystem,\n")
			g.w.W("Name: %s,\n", strconv.Quote("request_count"))
			g.w.W("Help: %s,\n", strconv.Quote("Number of requests received."))

			g.w.W("}, []string{\"method\", \"err\"")
			if len(g.Labels) > 0 {
				for _, l := range g.Labels {
					g.w.W(", %s", strconv.Quote(l.Name))
				}
			}
			g.w.W("})\n")

			g.w.W("\n}\n")

			g.w.W("if i.opts.requestLatency == nil {\n")
			g.w.W("i.opts.requestLatency = %s.NewSummaryFrom(%s.SummaryOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
			g.w.W("Namespace: namespace,\n")
			g.w.W("Subsystem: subsystem,\n")
			g.w.W("Name: %s,\n", strconv.Quote("request_latency_microseconds"))
			g.w.W("Help: %s,\n", strconv.Quote("Total duration of requests in microseconds."))
			g.w.W("}, []string{\"method\"")

			if len(g.Labels) > 0 {
				for _, l := range g.Labels {
					g.w.W(", %s", strconv.Quote(l.Name))
				}
			}

			g.w.W("})\n")
			g.w.W("\n}\n")
			g.w.W("return i\n}\n}\n")
		}
	}

	return g.w.Bytes()
}

func (g *Instrumenting) OutputPath() string {
	return ""
}

func (g *Instrumenting) Filename() string {
	return "instrumenting.go"
}
