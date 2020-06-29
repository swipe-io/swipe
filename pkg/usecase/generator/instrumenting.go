package generator

import (
	"context"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/pkg/types"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type instrumenting struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
	i        *importer.Importer
}

func (g *instrumenting) Process(ctx context.Context) error {
	metricsPkg := g.i.Import("metrics", "github.com/go-kit/kit/metrics")
	timePkg := g.i.Import("time", "time")
	typeStr := stdtypes.TypeString(g.o.Type, g.i.QualifyPkg)

	name := "instrumentingMiddleware" + g.o.ID
	constructName := "NewInstrumentingMiddleware" + g.o.ID

	g.W("type %s struct {\n", name)
	g.W("next %s\n", typeStr)
	g.W("requestCount %s.Counter\n", metricsPkg)
	g.W("requestLatency %s.Histogram\n", metricsPkg)
	g.W("}\n")

	for _, m := range g.o.Methods {
		var params []string

		if m.ParamCtx != nil {
			params = append(params, m.ParamCtx.Name(), stdtypes.TypeString(m.ParamCtx.Type(), g.i.QualifyPkg))
		}

		params = append(params, types.NameTypeParams(m.Params, g.i.QualifyPkg, nil)...)
		results := types.NameTypeParams(m.Results, g.i.QualifyPkg, nil)

		if m.ReturnErr != nil {
			results = append(results, "", "error")
		}

		g.WriteFunc(m.Name, "s *"+name, params, results, func() {
			g.WriteDefer(
				[]string{"begin " + timePkg + ".Time"},
				[]string{timePkg + ".Now()"},
				func() {
					g.W("s.requestCount.With(\"method\", \"%s\").Add(1)\n", m.Name)
					g.W("s.requestLatency.With(\"method\", \"%s\").Observe(%s.Since(begin).Seconds())\n", m.Name, timePkg)
				},
			)
			g.W("return s.next.%s(", m.Name)

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
	g.W("func %[1]s(s %[2]s, requestCount %[4]s.Counter, requestLatency %[4]s.Histogram) %[2]s {\n return &%[3]s{next: s, requestCount: requestCount, requestLatency: requestLatency}\n}\n", constructName, typeStr, name, metricsPkg)
	return nil
}

func (g *instrumenting) PkgName() string {
	return ""
}

func (g *instrumenting) OutputDir() string {
	return ""
}

func (g *instrumenting) Filename() string {
	return g.filename
}

func (g *instrumenting) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewInstrumenting(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &instrumenting{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
