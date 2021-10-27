package generator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Logging struct {
	w             writer.GoWriter
	Interfaces    []*config.Interface
	MethodOptions map[string]config.MethodOptions
}

func (g *Logging) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		loggerPkg := importer.Import("log", "github.com/go-kit/kit/log")
		levelPkg := importer.Import("level", "github.com/go-kit/kit/log/level")

		ifaceTypeName := NameInterface(iface)
		middlewareNameType := NameLoggingMiddleware(iface)
		middlewareFuncName := fmt.Sprintf("Logging%sMiddleware", UcNameWithAppPrefix(iface))
		middlewareTypeName := IfaceMiddlewareTypeName(iface)

		g.w.WriteTypeStruct(
			middlewareNameType,
			[]string{
				"next", ifaceTypeName,
				"logger", loggerPkg + ".Logger",
			},
		)

		for _, m := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+m.Name.Value]

			includes := map[string]struct{}{}
			excludes := map[string]struct{}{}

			for _, v := range mopt.LoggingParams.Includes {
				includes[v] = struct{}{}
			}
			for _, v := range mopt.LoggingParams.Excludes {
				excludes[v] = struct{}{}
			}

			logParams := makeLogParams(includes, excludes, m.Sig.Params...)

			if len(mopt.LoggingContext) > 0 {
				for _, lc := range mopt.LoggingContext {
					logParams = append(logParams, strconv.Quote(lc.Name), "ctx.Value("+swipe.TypeString(lc.Key, false, importer)+")")
				}
			}

			var params, results []string

			for _, param := range m.Sig.Params {
				var prefix string
				if param.IsVariadic {
					prefix = "..."
				}
				params = append(params, prefix+swipe.TypeString(param, false, importer))
			}

			for _, result := range m.Sig.Results {
				if IsError(result) {
					logParams = append(logParams, strconv.Quote("err"), result.Name.Value)
					continue
				}

				logParams = append(logParams, makeLogParam(result.Name.Value, result.Type)...)
				results = append(results, result.Name.Value, swipe.TypeString(result, false, importer))
			}

			g.w.W("func (s *%s) %s %s {\n", middlewareNameType, m.Name.Value, swipe.TypeString(m.Sig, false, importer))

			if mopt.Logging.Take() && len(logParams) > 0 {
				methodName := iface.Named.Name.Lower() + "." + m.Name.Value
				timePkg := importer.Import("time", "time")

				g.w.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					var resultErr *option.VarType
					for _, result := range m.Sig.Results {
						if IsError(result) {
							resultErr = result
							g.w.W("if logErr, ok := %s.(interface{LogError() error}); ok {\n", result.Name)
							g.w.W("%s = logErr.LogError()\n", result.Name)
							g.w.W("}\n")
						}
					}

					g.w.W("logger := %s.WithPrefix(s.logger, \"method\",\"%s\",\"took\",%s.Since(now))\n", loggerPkg, methodName, timePkg)
					if resultErr != nil {
						g.w.W("if %[2]s != nil {\nlogger = %[1]s.Error(logger)\n} else {\nlogger = %[1]s.Debug(logger)\n}\n", levelPkg, resultErr.Name)
					}
					g.w.W("_ = logger.Log(%s)\n", strings.Join(logParams, ","))
				})
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

			g.w.W("return")

			g.w.W("}\n")
		}

		g.w.W("func %[1]s(logger %[4]s.Logger) %[5]s {\nreturn func(next %[2]s) %[2]s {\nreturn &%[3]s{\nnext: next,\nlogger: logger,\n}\n}\n}\n", middlewareFuncName, ifaceTypeName, middlewareNameType, loggerPkg, middlewareTypeName)
	}
	return g.w.Bytes()
}

func (g *Logging) OutputDir() string {
	return ""
}

func (g *Logging) Filename() string {
	return "logging.go"
}
