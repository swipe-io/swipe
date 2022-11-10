package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/metric"
	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type Metric struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	Output     string
	Pkg        string
}

func (g *Metric) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	var interfaces []metric.Interface
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		metricInterface := metric.Interface{
			TypeName: NameInterface(iface),
			LcName:   LcNameWithAppPrefix(iface),
			UcName:   UcNameWithAppPrefix(iface),
		}
		for _, method := range ifaceType.Methods {
			metricInterface.Methods = append(metricInterface.Methods, metric.Method{
				Name:      method.Name,
				Sig:       method.Sig,
				IsEnabled: true,
			})
		}
		interfaces = append(interfaces, metricInterface)
	}

	data := metric.NewMetric(importer).
		SetInterfaces(interfaces).
		Build()
	_, _ = g.w.Write(data)

	return g.w.Bytes()
}

func (g *Metric) Package() string {
	return g.Pkg
}

func (g *Metric) OutputPath() string {
	return g.Output
}

func (g *Metric) Filename() string {
	return "metric.go"
}
