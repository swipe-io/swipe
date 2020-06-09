package service

import (
	"strconv"
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

	for _, m := range w.ctx.iface.methods {
		logParams := makeLogParams(m.params...)

		if len(m.results) > 0 {
			if m.resultsNamed {
				logParams = append(logParams, makeLogParams(m.results...)...)
			} else {
				logParams = append(logParams, strconv.Quote("result"), makeLogParam("result", m.results[0].Type()))
			}
		}

		var params, results []string

		if m.paramCtx != nil {
			params = append(params, m.paramCtx.Name(), w.w.TypeString(m.paramCtx.Type()))
		}

		params = append(params, utils.NameTypeParams(m.params, w.w.TypeString, nil)...)

		if len(m.results) > 0 {
			if m.resultsNamed {
				results = utils.NameTypeParams(m.results, w.w.TypeString, nil)
			} else {
				results = append(results, "result", w.w.TypeString(m.results[0].Type()))
			}
		}

		if m.returnErr != nil {
			errName := m.returnErr.Name()
			if errName == "" || errName == "_" {
				errName = "err"
			}
			results = append(results, errName, "error")

			logParams = append(logParams, strconv.Quote(errName), errName)
		}

		w.w.WriteFunc(m.name, "s *"+name, params, results, func() {
			if len(logParams) > 0 {
				w.w.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					w.w.Write("s.logger.Log(\"method\",\"%s\",\"took\",%s.Since(now),", m.name, timePkg)
					w.w.Write(strings.Join(logParams, ","))
					w.w.Write(")\n")
				})
			}
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

func newLogging(ctx serviceCtx, w *writer.Writer) *Logging {
	return &Logging{ctx: ctx, w: w}
}
