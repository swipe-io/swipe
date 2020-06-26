package processor

import (
	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	ug "github.com/swipe-io/swipe/pkg/usecase/generator"
)

type service struct {
	info   model.GenerateInfo
	option model.ServiceOption
}

func (p *service) SetOption(option interface{}) bool {
	o, ok := option.(model.ServiceOption)
	p.option = o
	return ok
}

func (p *service) Generators() []ug.Generator {
	generators := []ug.Generator{
		ug.NewEndpoint(p.info, p.option, importer.NewImporter(p.info.Pkg)),
	}
	if p.option.Transport.Protocol == "http" {
		generators = append(generators, ug.NewHttpTransport(p.info, p.option, importer.NewImporter(p.info.Pkg)))
		if p.option.Logging {
			generators = append(generators, ug.NewLogging(p.info, p.option, importer.NewImporter(p.info.Pkg)))
		}
		if p.option.Instrumenting.Enable {
			generators = append(generators, ug.NewInstrumenting(p.info, p.option, importer.NewImporter(p.info.Pkg)))
		}
		if p.option.Transport.JsonRPC.Enable {
			generators = append(generators, ug.NewJsonRPCServer(p.info, p.option, importer.NewImporter(p.info.Pkg)))
		} else {
			generators = append(generators, ug.NewRestServer(p.info, p.option, importer.NewImporter(p.info.Pkg)))
		}
		if p.option.Transport.Client.Enable {
			generators = append(generators, ug.NewClientStruct(p.info, p.option, importer.NewImporter(p.info.Pkg)))
			if p.option.Transport.JsonRPC.Enable {
				generators = append(
					generators,
					ug.NewJsonRPCGoClient(p.info, p.option, importer.NewImporter(p.info.Pkg)),
					ug.NewJsonRPCJSClient(p.info, p.option),
				)
			} else {
				generators = append(generators, ug.NewRestGoClient(p.info, p.option, importer.NewImporter(p.info.Pkg)))
			}
		}
	}
	if p.option.Transport.Openapi.Enable {
		generators = append(generators, ug.NewOpenapi(p.info, p.option))
	}
	return generators
}

func NewService(info model.GenerateInfo) Processor {
	return &service{info: info}
}
