package service

import (
	"go/types"
	"strings"

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

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)

		sig := m.Type().(*types.Signature)

		params := utils.NameTypeParams(sig.Params(), w.w.TypeString, nil)
		results := utils.NameTypeParams(sig.Results(), w.w.TypeString, nil)

		w.w.WriteFunc(m.Name(), "s *"+name, params, results, func() {
			w.w.WriteDefer(
				[]string{"begin " + timePkg + ".Time"},
				[]string{timePkg + ".Now()"},
				func() {
					w.w.Write("s.requestCount.With(\"method\", \"%s\").Add(1)\n", m.Name())
					w.w.Write("s.requestLatency.With(\"method\", \"%s\").Observe(%s.Since(begin).Seconds())\n", m.Name(), timePkg)
				},
			)
			w.w.Write("return s.next.%s(", m.Name())
			w.w.Write(strings.Join(utils.NameParams(sig.Params(), nil), ","))
			w.w.Write(")\n")
		})
	}
	return nil
}

func newInstrumenting(ctx serviceCtx, w *writer.Writer) *Instrumenting {
	return &Instrumenting{ctx: ctx, w: w}
}
