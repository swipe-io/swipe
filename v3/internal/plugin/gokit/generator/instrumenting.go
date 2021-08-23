package generator

import (
	"context"
	"fmt"
	"strconv"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Instrumenting struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodDefaultOption
}

func (g *Instrumenting) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	metricsPkg := importer.Import("metrics", "github.com/go-kit/kit/metrics")

	g.w.W("type instrumentingOpts struct {\n")
	g.w.W("requestCount %s.Counter\n", metricsPkg)
	g.w.W("requestLatency %s.Histogram\n", metricsPkg)
	g.w.W("namespace string\n")
	g.w.W("subsystem string\n")
	g.w.W("}\n\n")

	g.w.W("type InstrumentingOption func(*instrumentingOpts)\n\n")

	g.w.W("func Namespace(v string) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.namespace = v\n}\n}\n\n")
	g.w.W("func Subsystem(v string) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.subsystem = v\n}\n}\n\n")

	g.w.W("func RequestLatency(requestLatency %s.Histogram) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.requestLatency = requestLatency\n}\n}\n\n", metricsPkg)
	g.w.W("func RequestCount(requestCount %s.Counter) InstrumentingOption {\nreturn func(o *instrumentingOpts) {\no.requestCount = requestCount\n}\n}\n\n", metricsPkg)

	if len(g.Interfaces) > 0 {
		timePkg := importer.Import("time", "time")
		stdPrometheusPkg := importer.Import("prometheus", "github.com/prometheus/client_golang/prometheus")
		kitPrometheusPkg := importer.Import("prometheus", "github.com/go-kit/kit/metrics/prometheus")

		for _, iface := range g.Interfaces {
			ifaceType := iface.Named.Type.(*option.IfaceType)

			ifaceName := NameInterface(iface)
			name := NameInstrumentingMiddleware(iface)

			constructName := fmt.Sprintf("NewInstrumenting%sMiddleware", UcNameWithAppPrefix(iface))

			g.w.W("type %s struct {\n", name)
			g.w.W("next %s\n", ifaceName)
			g.w.W("opts *instrumentingOpts\n")
			g.w.W("}\n\n")

			for _, m := range ifaceType.Methods {
				mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

				g.w.W("func (s *%s) %s %s {\n", name, m.Name.Value, importer.TypeString(m.Sig))
				if mopt.Instrumenting.Take() {
					methodName := iface.Named.Name.Lower() + "." + m.Name.Value
					g.w.WriteDefer(
						[]string{"begin " + timePkg + ".Time"},
						[]string{timePkg + ".Now()"},
						func() {
							e := Error(m.Sig.Results)
							if e != nil {
								g.w.W("if %[1]s != nil {\ns.opts.requestCount.With(\"method\", \"%[2]s\", \"err\", %[1]s.Error()).Add(1)\n} else {\n", e.Name, methodName)
								g.w.W("")
							}
							g.w.W("s.opts.requestCount.With(\"method\", \"%s\", \"err\", \"\").Add(1)\n", methodName)
							if e != nil {
								g.w.W("}\n")
							}
							g.w.W("s.opts.requestLatency.With(\"method\", \"%s\").Observe(%s.Since(begin).Seconds())\n", methodName, timePkg)
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

			g.w.W("func %[1]s(s %[2]s, opts ...InstrumentingOption) *%[3]s {\n", constructName, ifaceName, name)
			g.w.W("i := &%s{next: s, opts: &instrumentingOpts{}}\n", name)

			g.w.W("for _, o := range opts {\no(i.opts)\n}\n")

			g.w.W("if i.opts.requestCount == nil {\n")
			g.w.W("i.opts.requestCount = %s.NewCounterFrom(%s.CounterOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
			g.w.W("Namespace: i.opts.namespace,\n")
			g.w.W("Subsystem: i.opts.subsystem,\n")
			g.w.W("Name: %s,\n", strconv.Quote("request_count"))
			g.w.W("Help: %s,\n", strconv.Quote("Number of requests received."))
			g.w.W("}, []string{\"method\", \"err\"})\n")
			g.w.W("\n}\n")

			g.w.W("if i.opts.requestLatency == nil {\n")
			g.w.W("i.opts.requestLatency = %s.NewSummaryFrom(%s.SummaryOpts{\n", kitPrometheusPkg, stdPrometheusPkg)
			g.w.W("Namespace: i.opts.namespace,\n")
			g.w.W("Subsystem: i.opts.subsystem,\n")
			g.w.W("Name: %s,\n", strconv.Quote("request_latency_microseconds"))
			g.w.W("Help: %s,\n", strconv.Quote("Total duration of requests in microseconds."))
			g.w.W("}, []string{\"method\"})\n")
			g.w.W("\n}\n")

			g.w.W("return i\n}\n")
		}
	}

	return g.w.Bytes()
}

func (g *Instrumenting) OutputDir() string {
	return ""
}

func (g *Instrumenting) Filename() string {
	return "instrumenting.go"
}
