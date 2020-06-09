package service

import (
	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Instrumenting struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *Instrumenting) Write() error {
	metricsPkg := w.w.Import("metrics", "github.com/go-kit/kit/metrics")
	timePkg := w.w.Import("time", "time")

	name := "instrumentingMiddleware" + w.ctx.id

	w.w.Write("type %s struct {\n", name)
	w.w.Write("next %s\n", w.ctx.typeStr)
	w.w.Write("requestCount %s.Counter\n", metricsPkg)
	w.w.Write("requestLatency %s.Histogram\n", metricsPkg)
	w.w.Write("}\n")

	for _, m := range w.ctx.iface.methods {
		var params []string

		if m.paramCtx != nil {
			params = append(params, m.paramCtx.Name(), w.w.TypeString(m.paramCtx.Type()))
		}

		params = append(params, utils.NameTypeParams(m.params, w.w.TypeString, nil)...)
		results := utils.NameTypeParams(m.results, w.w.TypeString, nil)

		if m.returnErr != nil {
			results = append(results, "", "error")
		}

		w.w.WriteFunc(m.name, "s *"+name, params, results, func() {
			w.w.WriteDefer(
				[]string{"begin " + timePkg + ".Time"},
				[]string{timePkg + ".Now()"},
				func() {
					w.w.Write("s.requestCount.With(\"method\", \"%s\").Add(1)\n", m.name)
					w.w.Write("s.requestLatency.With(\"method\", \"%s\").Observe(%s.Since(begin).Seconds())\n", m.name, timePkg)
				},
			)
			w.w.Write("return s.next.%s(", m.name)

			if m.paramCtx != nil {
				w.w.Write("%s,", m.paramCtx.Name())
			}

			for i, p := range m.params {
				if i > 0 {
					w.w.Write(",")
				}
				w.w.Write(p.Name())
			}

			w.w.Write(")\n")
		})
	}
	return nil
}

func newInstrumenting(ctx serviceCtx, w *writer.Writer) *Instrumenting {
	return &Instrumenting{ctx: ctx, w: w}
}
