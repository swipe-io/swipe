package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/logging"
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

	var interfaces []logging.Interface
	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		loggingInterface := logging.Interface{
			TypeName: NameInterface(iface),
			LcName:   LcNameWithAppPrefix(iface),
			UcName:   UcNameWithAppPrefix(iface),
		}
		for _, method := range ifaceType.Methods {
			mopt := g.MethodOptions[iface.Named.Name.Value+method.Name.Value]
			loggingMethod := logging.Method{
				Name: method.Name,
				Sig:  method.Sig,
			}
			for _, loggingContext := range mopt.LoggingContext {
				loggingMethod.ParamsContexts = append(loggingMethod.ParamsContexts, logging.ParamContext{
					Key:  loggingContext.Key,
					Name: loggingContext.Name,
				})
			}
			for _, param := range mopt.LoggingParams.Includes {
				loggingMethod.ParamsIncludes = append(loggingMethod.ParamsIncludes, param)
			}
			for _, param := range mopt.LoggingParams.Excludes {
				loggingMethod.ParamsExcludes = append(loggingMethod.ParamsExcludes, param)
			}
			loggingInterface.Methods = append(loggingInterface.Methods, loggingMethod)
		}
		interfaces = append(interfaces, loggingInterface)
	}

	data := logging.NewLogging(importer).SetInterfaces(interfaces).Build()
	_, _ = g.w.Write(data)

	return g.w.Bytes()
}

func (g *Logging) OutputPath() string {
	return ""
}

func (g *Logging) Filename() string {
	return "logging.go"
}
