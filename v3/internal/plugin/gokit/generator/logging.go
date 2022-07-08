package generator

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/v3/internal/plugin"

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

	loggerPkg := importer.Import("log", "github.com/go-kit/log")
	levelPkg := importer.Import("level", "github.com/go-kit/log/level")

	g.w.W("type errLevel interface {\n\tLevel() string\n}\n\n")

	g.w.W("func levelLogger(e errLevel, logger %[1]s.Logger) %[1]s.Logger {\n", loggerPkg)
	g.w.W("switch e.Level() {\n")
	g.w.W("default:\nreturn %s.Error(logger)\n", levelPkg)
	g.w.W("case \"debug\":\nreturn %s.Debug(logger)\n", levelPkg)
	g.w.W("case \"info\":\nreturn %s.Info(logger)\n", levelPkg)
	g.w.W("case \"warn\":\nreturn %s.Warn(logger)\n", levelPkg)
	g.w.W("}\n")
	g.w.W("}\n\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

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

			var (
				logResults []string
				errorVars  []*option.VarType
			)

			for _, result := range m.Sig.Results {
				if plugin.IsError(result) {

					errorVars = append(errorVars, result)
					continue
				}
				logResults = append(logResults, makeLogParam(result.Name.Value, result.Type)...)
			}

			g.w.W("func (s *%s) %s %s {\n", middlewareNameType, m.Name.Value, swipe.TypeString(m.Sig, false, importer))

			if mopt.Logging.Take() && (len(logParams) > 0 || len(logResults) > 0 || len(errorVars) > 0) {
				methodName := iface.Named.Name.Lower() + "." + m.Name.Value
				timePkg := importer.Import("time", "time")

				g.w.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					//var resultErr *option.VarType
					for _, result := range m.Sig.Results {
						if plugin.IsError(result) {
							//resultErr = result
							g.w.W("if logErr, ok := %s.(interface{LogError() error}); ok {\n", result.Name)
							g.w.W("%s = logErr.LogError()\n", result.Name)
							g.w.W("}\n")
						}
					}
					g.w.W("logger := %s.WithPrefix(s.logger, \"method\",\"%s\",\"took\",%s.Since(now))\n", loggerPkg, methodName, timePkg)

					logParamsStr := strings.Join(logParams, ",")
					logResultsStr := strings.Join(logResults, ",")

					if logParamsStr != "" {
						logResultsStr = "," + logResultsStr
					}

					if len(errorVars) > 0 {
						for _, errorVar := range errorVars {
							errLogStr := fmt.Sprintf("%s, %s", strconv.Quote(errorVar.Name.String()), errorVar.Name.String())
							if logParamsStr != "" {
								errLogStr = "," + errLogStr
							}
							g.w.W("if %s != nil {\n", errorVar.Name)
							g.w.W("if e, ok := %s.(errLevel); ok {\n", errorVar.Name)
							g.w.W("logger = levelLogger(e, logger)\n")
							g.w.W("} else {\n_ = %s.Error(logger).Log(%s)\n}\n", levelPkg, logParamsStr+errLogStr)
							g.w.W("} else {\n_ = %s.Debug(logger).Log(%s)\n}\n", levelPkg, logParamsStr+logResultsStr)
						}
					} else {
						g.w.W("_ = %s.Debug(logger).Log(%s)\n", levelPkg, logParamsStr+logResultsStr)
					}
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

			g.w.W("return\n")

			g.w.W("}\n")
		}

		g.w.W("func %[1]s(logger %[4]s.Logger) %[5]s {\nreturn func(next %[2]s) %[2]s {\nreturn &%[3]s{\nnext: next,\nlogger: logger,\n}\n}\n}\n", middlewareFuncName, ifaceTypeName, middlewareNameType, loggerPkg, middlewareTypeName)
	}
	return g.w.Bytes()
}

func (g *Logging) OutputPath() string {
	return ""
}

func (g *Logging) Filename() string {
	return "logging.go"
}
