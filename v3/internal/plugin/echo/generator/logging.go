package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/writer"

	"github.com/swipe-io/swipe/v3/option"

	"github.com/swipe-io/swipe/v3/swipe"

	"github.com/swipe-io/swipe/v3/internal/logging"

	"github.com/swipe-io/swipe/v3/internal/plugin/echo/config"
)

type Logging struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	Output     string
	Pkg        string
}

func (g *Logging) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	var interfaces []logging.Interface
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		loggingInterface := logging.Interface{
			TypeName: NameInterface(iface),
			LcName:   LcNameWithAppPrefix(iface),
			UcName:   UcNameWithAppPrefix(iface),
		}
		for _, method := range ifaceType.Methods {
			loggingInterface.Methods = append(loggingInterface.Methods, logging.Method{
				Name:           method.Name,
				Sig:            method.Sig,
				ParamsIncludes: nil,
				ParamsExcludes: nil,
				ParamsContexts: nil,
			})
		}
		interfaces = append(interfaces, loggingInterface)
	}

	data := logging.NewLogging(importer).SetInterfaces(interfaces).Build()
	_, _ = g.w.Write(data)

	return g.w.Bytes()
}

func (g *Logging) Package() string {
	return g.Pkg
}

func (g *Logging) OutputPath() string {
	return g.Output
}

func (g *Logging) Filename() string {
	return "logging.go"
}
