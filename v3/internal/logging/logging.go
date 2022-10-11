package logging

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/v3/internal/plugin"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type ParamContext struct {
	Key  interface{}
	Name string
}

type Interface struct {
	TypeName string
	LcName   string
	UcName   string
	Methods  []Method
}

type Method struct {
	Name           option.String
	Sig            *option.SignType
	ParamsIncludes []string
	ParamsExcludes []string
	ParamsContexts []ParamContext
}

type Logging struct {
	w          writer.GoWriter
	importer   swipe.Importer
	interfaces []Interface
}

func (l *Logging) SetInterfaces(interfaces []Interface) *Logging {
	l.interfaces = interfaces
	return l
}

func (l *Logging) Build() []byte {
	loggerPkg := l.importer.Import("log", "github.com/go-kit/log")
	levelPkg := l.importer.Import("level", "github.com/go-kit/log/level")

	l.w.W("type errLevel interface {\n\tLevel() string\n}\n\n")
	l.w.W("func levelLogger(e errLevel, logger %[1]s.Logger) %[1]s.Logger {\n", loggerPkg)
	l.w.W("switch e.Level() {\n")
	l.w.W("default:\nreturn %s.Error(logger)\n", levelPkg)
	l.w.W("case \"debug\":\nreturn %s.Debug(logger)\n", levelPkg)
	l.w.W("case \"info\":\nreturn %s.Info(logger)\n", levelPkg)
	l.w.W("case \"warn\":\nreturn %s.Warn(logger)\n", levelPkg)
	l.w.W("}\n")
	l.w.W("}\n\n")

	for _, iface := range l.interfaces {

		middlewareNameType := iface.LcName + "LoggingMiddleware"
		middlewareFuncName := fmt.Sprintf("Logging%sMiddleware", iface.UcName)
		middlewareTypeName := iface.UcName + "Middleware"

		l.w.WriteTypeStruct(
			middlewareNameType,
			[]string{
				"next", iface.TypeName,
				"logger", loggerPkg + ".Logger",
			},
		)
		for _, m := range iface.Methods {
			includes := map[string]struct{}{}
			excludes := map[string]struct{}{}

			for _, v := range m.ParamsIncludes {
				includes[v] = struct{}{}
			}
			for _, v := range m.ParamsExcludes {
				excludes[v] = struct{}{}
			}

			logParams := makeLogParams(includes, excludes, m.Sig.Params...)

			if len(m.ParamsContexts) > 0 {
				for _, lc := range m.ParamsContexts {
					logParams = append(logParams, strconv.Quote(lc.Name), "ctx.Value("+swipe.TypeString(lc.Key, false, l.importer)+")")
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

			l.w.W("func (s *%s) %s %s {\n", middlewareNameType, m.Name.Value, swipe.TypeString(m.Sig, false, l.importer))

			if len(logParams) > 0 || len(logResults) > 0 || len(errorVars) > 0 {
				methodName := m.Name.Value
				timePkg := l.importer.Import("time", "time")

				l.w.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					for _, result := range m.Sig.Results {
						if plugin.IsError(result) {
							l.w.W("if logErr, ok := %s.(interface{LogError() error}); ok {\n", result.Name)
							l.w.W("%s = logErr.LogError()\n", result.Name)
							l.w.W("}\n")
						}
					}
					l.w.W("logger := %[1]s.WithPrefix(s.logger, \"message\", \"call method - %[2]s\", \"method\",\"%[2]s\",\"took\",%[3]s.Since(now))\n", loggerPkg, methodName, timePkg)

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
							l.w.W("if %s != nil {\n", errorVar.Name)
							l.w.W("if e, ok := %s.(errLevel); ok {\n", errorVar.Name)
							l.w.W("logger = levelLogger(e, logger)\n")
							l.w.W("} else {\n_ = %s.Error(logger).Log(%s)\n}\n", levelPkg, logParamsStr+errLogStr)
							l.w.W("} else {\n_ = %s.Debug(logger).Log(%s)\n}\n", levelPkg, logParamsStr+logResultsStr)
						}
					} else {
						l.w.W("_ = %s.Debug(logger).Log(%s)\n", levelPkg, logParamsStr+logResultsStr)
					}
				})
			}

			if len(m.Sig.Results) > 0 {
				for i, result := range m.Sig.Results {
					if i > 0 {
						l.w.W(",")
					}
					l.w.W(result.Name.Value)
				}
				l.w.W(" = ")
			}

			l.w.W("s.next.%s(", m.Name)
			for i, param := range m.Sig.Params {
				if i > 0 {
					l.w.W(",")
				}
				var variadic string
				if param.IsVariadic {
					variadic = "..."
				}
				l.w.W(param.Name.Value + variadic)
			}
			l.w.W(")\n")

			l.w.W("return\n")

			l.w.W("}\n")
		}
		l.w.W("func %[1]s(logger %[4]s.Logger) %[5]s {\nreturn func(next %[2]s) %[2]s {\nreturn &%[3]s{\nnext: next,\nlogger: logger,\n}\n}\n}\n", middlewareFuncName, iface.TypeName, middlewareNameType, loggerPkg, middlewareTypeName)
	}

	return l.w.Bytes()
}

func NewLogging(importer swipe.Importer) *Logging {
	return &Logging{importer: importer}
}
