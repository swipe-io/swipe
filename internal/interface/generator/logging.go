package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/strcase"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/types"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type loggingGeneratorOptionsGateway interface {
	AppID() string
	Interfaces() model.Interfaces
	MethodOption(m model.ServiceMethod) model.MethodOption
}

type logging struct {
	writer.GoLangWriter
	options loggingGeneratorOptionsGateway
	i       *importer.Importer
}

func (g *logging) Prepare(ctx context.Context) error {
	return nil
}

func (g *logging) Process(ctx context.Context) error {
	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		timePkg := g.i.Import("time", "time")
		loggerPkg := g.i.Import("log", "github.com/go-kit/kit/log")
		typeStr := stdtypes.TypeString(iface.Type(), g.i.QualifyPkg)

		name := iface.Name() + "LoggingMiddleware"
		constructName := fmt.Sprintf("NewLogging%sMiddleware", iface.Name())

		g.WriteTypeStruct(
			name,
			[]string{
				"next", typeStr,
				"logger", loggerPkg + ".Logger",
			},
		)

		for _, m := range iface.Methods() {
			mopt := g.options.MethodOption(m)

			logParams := makeLogParams(mopt.LoggingIncludeParams, mopt.LoggingExcludeParams, m.Params...)

			if len(m.Results) > 0 {
				if m.ResultsNamed {
					logParams = append(logParams, makeLogParams(mopt.LoggingIncludeParams, mopt.LoggingExcludeParams, m.Results...)...)
				} else {
					if paramName := makeLogParam("result", m.Results[0].Type()); paramName != nil {
						logParams = append(logParams, paramName...)
					}
				}
			}

			var params, results []string

			if m.ParamCtx != nil {
				params = append(params, m.ParamCtx.Name(), stdtypes.TypeString(m.ParamCtx.Type(), g.i.QualifyPkg))
			}

			params = append(params, types.NameTypeParams(m.Params, g.i.QualifyPkg, nil)...)

			if len(m.Results) > 0 {
				if m.ResultsNamed {
					results = types.NameType(m.Results, g.i.QualifyPkg, nil)
				} else {
					results = append(results, "", stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg))
				}
			}

			if m.ReturnErr != nil {
				results = append(results, "", "error")
				logParams = append(logParams, strconv.Quote("err"), "logErr")
			}

			g.WriteFunc(m.Name, "s *"+name, params, results, func() {
				if m.ReturnErr != nil || len(m.Results) > 0 {
					g.WriteVarGroup(func() {
						for _, result := range m.Results {
							name := "result"
							if m.ResultsNamed {
								name = strcase.ToLowerCamel(result.Name())
							}
							g.W("%s %s\n", name, stdtypes.TypeString(result.Type(), g.i.QualifyPkg))
						}
						if m.ReturnErr != nil {
							g.W("err error\n")
						}
					})
				}

				if mopt.LoggingEnable {
					if len(logParams) > 0 {
						g.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
							if m.ReturnErr != nil {
								g.W("logErr := err\n")
								g.W("if le, ok := err.(interface{LogError() error}); ok {\n")
								g.W("logErr = le.LogError()\n")
								g.W("}\n")
							}
							g.W("s.logger.Log(\"method\",\"%s\",\"took\",%s.Since(now),", m.Name, timePkg)
							g.W(strings.Join(logParams, ","))
							g.W(")\n")
						})
					}

				}

				if len(m.Results) > 0 || m.ReturnErr != nil {
					for i, result := range m.Results {
						name := "result"
						if m.ResultsNamed {
							name = strcase.ToLowerCamel(result.Name())
						}
						if i > 0 {
							g.W(",")
						}
						g.W(name)
					}
					if m.ReturnErr != nil {
						if len(m.Results) > 0 {
							g.W(",")
						}
						g.W("err")
					}

					g.W(" = ")
				}

				g.W("s.next.%s(", m.Name)
				if m.ParamCtx != nil {
					g.W("%s,", m.ParamCtx.Name())
				}
				for i, p := range m.Params {
					if i > 0 {
						g.W(",")
					}
					g.W(p.Name())
				}
				g.W(")\n")

				if len(m.Results) > 0 || m.ReturnErr != nil {
					generateNameResult := func(format string, args ...interface{}) {
						for i, result := range m.Results {
							name := "result"
							if m.ResultsNamed {
								name = strcase.ToLowerCamel(result.Name())
							}
							if i > 0 {
								g.W(",")
							}
							g.W(name)
						}
						if m.ReturnErr != nil {
							if len(m.Results) > 0 {
								g.W(",")
							}
							g.W(format, args...)
						}
					}

					g.W("return ")

					generateNameResult("err")
				}
			})
		}

		g.W("func %[1]s(s %[2]s, logger %[4]s.Logger) %[2]s {\n return &%[3]s{next: s, logger: logger}\n}\n", constructName, typeStr, name, loggerPkg)
	}
	return nil
}

func (g *logging) PkgName() string {
	return ""
}

func (g *logging) OutputDir() string {
	return ""
}

func (g *logging) Filename() string {
	return "logging_gen.go"
}

func (g *logging) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewLogging(options loggingGeneratorOptionsGateway) generator.Generator {
	return &logging{
		options: options,
	}
}
