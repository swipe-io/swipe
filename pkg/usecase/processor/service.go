package processor

import (
	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	ug "github.com/swipe-io/swipe/pkg/usecase/generator"
)

type service struct {
	info      model.GenerateInfo
	option    model.ServiceOption
	importers map[string]*importer.Importer
}

func (p *service) SetOption(option interface{}) bool {
	o, ok := option.(model.ServiceOption)
	p.option = o
	return ok
}

func (p *service) Generators() []ug.Generator {
	generators := []ug.Generator{
		ug.NewEndpoint("endpoint_gen.go", p.info, p.option),
		ug.NewEndpointFactory("endpoint_gen.go", p.info, p.option),
	}
	if p.option.Transport.Protocol == "http" {
		generators = append(generators, ug.NewHttpTransport("http_gen.go", p.info, p.option))
		if p.option.Logging {
			generators = append(generators, ug.NewLogging("logging_gen.go", p.info, p.option))
		}
		if p.option.Instrumenting.Enable {
			generators = append(generators, ug.NewInstrumenting("instrumenting_gen.go", p.info, p.option))
		}
		if p.option.Transport.JsonRPC.Enable {
			generators = append(generators, ug.NewJsonRPCServer("server_gen.go", p.info, p.option))
		} else {
			generators = append(generators, ug.NewRestServer("server_gen.go", p.info, p.option))
		}
		if p.option.Transport.Client.Enable {
			generators = append(generators, ug.NewClientStruct("client_gen.go", p.info, p.option))
			if p.option.Transport.JsonRPC.Enable {
				generators = append(
					generators,
					ug.NewJsonRPCGoClient("client_gen.go", p.info, p.option),
					ug.NewJsonRPCJSClient("client_jsonrpc_gen.js", p.info, p.option),
				)
			} else {
				generators = append(generators, ug.NewRestGoClient("client_gen.go", p.info, p.option))
			}
		}
	}
	if p.option.Transport.Openapi.Enable {
		generators = append(generators, ug.NewOpenapi(p.info, p.option))
	}

	return generators
}

func NewService(info model.GenerateInfo) Processor {
	return &service{info: info, importers: map[string]*importer.Importer{}}
}
