package service

import (
	"go/types"
	"strings"

	"github.com/swipe-io/swipe/pkg/utils"
	"github.com/swipe-io/swipe/pkg/writer"
)

type Logging struct {
	ctx serviceCtx
	w   *writer.Writer
}

func (w *Logging) Write() error {
	loggerPkg := w.w.Import("log", "github.com/go-kit/kit/log")
	timePkg := w.w.Import("time", "time")

	name := "loggingMiddleware" + w.ctx.id

	w.w.WriteTypeStruct(
		name,
		[]string{
			"next", w.ctx.typeStr,
			"logger", loggerPkg + ".Logger",
		},
	)

	for i := 0; i < w.ctx.iface.NumMethods(); i++ {
		m := w.ctx.iface.Method(i)
		sig := m.Type().(*types.Signature)

		logParams := logParams(sig.Params(), sig.Results())

		params := utils.NameTypeParams(sig.Params(), w.w.TypeString, nil)
		results := utils.NameTypeParams(sig.Results(), w.w.TypeString, nil)

		w.w.WriteFunc(m.Name(), "s *"+name, params, results, func() {
			if len(logParams) > 0 {
				w.w.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					w.w.Write("s.logger.Log(\"method\",\"%s\",\"took\",%s.Since(now),", m.Name(), timePkg)
					w.w.Write(strings.Join(logParams, ","))
					w.w.Write(")\n")
				})
			}
			w.w.Write("return s.next.%s(", m.Name())
			w.w.Write(strings.Join(utils.NameParams(sig.Params(), nil), ","))
			w.w.Write(")\n")
		})
	}
	return nil
}

func newLogging(ctx serviceCtx, w *writer.Writer) *Logging {
	return &Logging{ctx: ctx, w: w}
}
