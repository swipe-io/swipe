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
	MethodOptions map[string]config.MethodOption
}

func (g *Logging) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)

		loggerPkg := importer.Import("log", "github.com/go-kit/kit/log")
		levelPkg := importer.Import("level", "github.com/go-kit/kit/log/level")

		ifaceTypeName := NameInterface(iface)
		name := NameLoggingMiddleware(iface)
		constructName := fmt.Sprintf("NewLogging%sMiddleware", UcNameWithAppPrefix(iface))

		g.w.WriteTypeStruct(
			name,
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
					logParams = append(logParams, strconv.Quote(lc.Name), "ctx.Value("+importer.TypeString(lc.Key)+")")
				}
			}

			var params, results []string

			for _, param := range m.Sig.Params {
				var prefix string
				if param.IsVariadic {
					prefix = "..."
				}
				params = append(params, prefix+importer.TypeString(param))
			}

			for _, result := range m.Sig.Results {
				if IsError(result) {
					logParams = append(logParams, strconv.Quote("err"), result.Name.Value)
					continue
				}
				logParams = append(logParams, strconv.Quote(result.Name.Value), result.Name.Value)
				results = append(results, result.Name.Value, importer.TypeString(result))
			}

			g.w.W("func (s *%s) %s %s {\n", name, m.Name.Value, importer.TypeString(m.Sig))

			if mopt.Logging.Value && len(logParams) > 0 {
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
		g.w.W("func %[1]s(s %[2]s, logger %[4]s.Logger) *%[3]s {\n return &%[3]s{next: s, logger: logger}\n}\n", constructName, ifaceTypeName, name, loggerPkg)
	}
	return g.w.Bytes()
}

func (g *Logging) OutputDir() string {
	return ""
}

func (g *Logging) Filename() string {
	return "logging.go"
}
