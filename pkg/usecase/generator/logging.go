package generator

import (
	"context"
	stdtypes "go/types"
	"strconv"
	"strings"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/types"
	"github.com/swipe-io/swipe/pkg/writer"
)

type logging struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
	i        *importer.Importer
}

func (g *logging) Prepare(ctx context.Context) error {
	return nil
}

func (g *logging) Process(ctx context.Context) error {
	var (
		timePkg string
	)
	if len(g.o.Methods) > 0 {
		timePkg = g.i.Import("time", "time")
	}
	loggerPkg := g.i.Import("log", "github.com/go-kit/kit/log")
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	name := "loggingMiddleware" + g.o.ID
	constructName := "NewLoggingMiddleware" + g.o.ID

	g.WriteTypeStruct(
		name,
		[]string{
			"next", typeStr,
			"logger", loggerPkg + ".Logger",
		},
	)

	for _, m := range g.o.Methods {
		logParams := makeLogParams(m.Params...)

		if len(m.Results) > 0 {
			if m.ResultsNamed {
				logParams = append(logParams, makeLogParams(m.Results...)...)
			} else {
				logParams = append(logParams, strconv.Quote("result"), makeLogParam("result", m.Results[0].Type()))
			}
		}

		var params, results []string

		if m.ParamCtx != nil {
			params = append(params, m.ParamCtx.Name(), stdtypes.TypeString(m.ParamCtx.Type(), g.i.QualifyPkg))
		}

		params = append(params, types.NameTypeParams(m.Params, g.i.QualifyPkg, nil)...)

		if len(m.Results) > 0 {
			if m.ResultsNamed {
				results = types.NameTypeParams(m.Results, g.i.QualifyPkg, nil)
			} else {
				results = append(results, "result", stdtypes.TypeString(m.Results[0].Type(), g.i.QualifyPkg))
			}
		}

		if m.ReturnErr != nil {
			errName := m.ReturnErr.Name()
			if errName == "" || errName == "_" {
				errName = "err"
			}
			results = append(results, errName, "error")

			logParams = append(logParams, strconv.Quote(errName), errName)
		}

		g.WriteFunc(m.Name, "s *"+name, params, results, func() {
			if len(logParams) > 0 {
				g.WriteDefer([]string{"now " + timePkg + ".Time"}, []string{timePkg + ".Now()"}, func() {
					g.W("s.logger.Log(\"method\",\"%s\",\"took\",%s.Since(now),", m.Name, timePkg)
					g.W(strings.Join(logParams, ","))
					g.W(")\n")
				})
			}
			if len(m.Results) > 0 {
				g.W("return ")
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
		})
	}

	g.W("func %[1]s(s %[2]s, logger %[4]s.Logger) %[2]s {\n return &%[3]s{next: s, logger: logger}\n}\n", constructName, typeStr, name, loggerPkg)
	return nil
}

func (g *logging) PkgName() string {
	return ""
}

func (g *logging) OutputDir() string {
	return ""
}

func (g *logging) Filename() string {
	return g.filename
}

func (g *logging) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewLogging(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &logging{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
