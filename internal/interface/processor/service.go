package processor

import (
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/git"
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
)

type serviceProcessor struct {
	sg      gateway.ServiceGateway
	gi      *git.GIT
	workDir string
	pkg     *packages.Package
}

func (p *serviceProcessor) Pkg() *packages.Package {
	return p.pkg
}

func (p *serviceProcessor) Generators() []ug.Generator {
	var generators []ug.Generator
	if p.sg.FoundService() {
		generators = append(generators, generator.NewEndpoint(p.sg))
	}
	if p.sg.FoundServiceGateway() {
		generators = append(
			generators,
			generator.NewGatewayGenerator(p.sg.Interfaces()),
		)
	}
	if p.sg.ReadmeEnable() {
		tags, _ := p.gi.GetTags()
		generators = append(generators,
			generator.NewReadme(
				p.sg,
				p.pkg.PkgPath,
				p.workDir,
				tags,
			),
		)
	}
	if p.sg.TransportType() == model.HTTPTransport {
		generators = append(generators, generator.NewHttpTransport(p.sg))
		if p.sg.LoggingEnable() {
			generators = append(generators, generator.NewLogging(p.sg))
		}
		if p.sg.InstrumentingEnable() {
			generators = append(generators, generator.NewInstrumenting(p.sg))
		}
		if p.sg.JSONRPCEnable() {
			if p.sg.JSONRPCDocEnable() {
				generators = append(generators, generator.NewJsonrpcDoc(p.sg, p.workDir))
			}
			generators = append(generators, generator.NewJsonRPCServer(p.sg))
		} else {
			generators = append(generators, generator.NewRestServer(p.sg))
		}
	}
	if p.sg.ClientEnable() {
		if p.sg.GoClientEnable() {
			generators = append(generators,
				generator.NewClientStruct(p.sg),
				generator.NewEndpointFactory(p.sg.Interfaces(), p.sg.Prefix()),
			)
		}
		if p.sg.JSONRPCEnable() {
			if p.sg.GoClientEnable() {
				generators = append(
					generators,
					generator.NewJsonRPCGoClient(p.sg),
				)
			}
			if p.sg.JSClientEnable() {
				generators = append(
					generators,
					generator.NewJsonRPCJSClient(p.sg),
				)
			}
		} else if p.sg.GoClientEnable() {
			generators = append(generators, generator.NewRestGoClient(p.sg))
		}
	}
	if p.sg.OpenapiEnable() {
		generators = append(generators, generator.NewOpenapi(p.sg, p.workDir))
	}
	return generators
}

func NewService(
	sg gateway.ServiceGateway,
	gi *git.GIT,
	workDir string,
	pkg *packages.Package,
) processor.Processor {
	return &serviceProcessor{
		sg:      sg,
		gi:      gi,
		workDir: workDir,
		pkg:     pkg,
	}
}
