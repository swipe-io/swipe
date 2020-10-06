package processor

import (
	"github.com/swipe-io/swipe/v2/internal/git"
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
)

type serviceProcessor struct {
	sg       gateway.ServiceGateway
	gi       *git.GIT
	workDir  string
	comments *typeutil.Map
	enums    *typeutil.Map
	pkg      *packages.Package
}

func (p *serviceProcessor) Pkg() *packages.Package {
	return p.pkg
}

func (p *serviceProcessor) Generators() []ug.Generator {
	generators := []ug.Generator{
		generator.NewEndpoint(p.sg.ID(), p.sg.Type(), p.sg.Methods()),
	}
	if p.sg.Readme().Enable {
		tags, _ := p.gi.GetTags()
		generators = append(generators,
			generator.NewReadme(
				p.sg.ID(),
				p.sg.RawID(),
				p.pkg.PkgPath,
				p.workDir,
				p.sg.Transport(),
				p.sg.Readme(),
				tags,
			),
		)
	}
	if p.sg.Transport().MarkdownDoc.Enable {
		generators = append(generators, generator.NewJsonrpcMarkdownDoc(
			p.sg.ID(),
			p.sg.Methods(),
			p.sg.Transport(),
			p.comments,
			p.enums,
			p.workDir,
		))
	}
	if p.sg.Transport().Protocol == "http" {
		generators = append(generators, generator.NewHttpTransport(p.sg.ID(), p.sg.Methods(), p.sg.Transport(), p.sg.Errors()))
		if p.sg.LoggingEnable() {
			generators = append(generators, generator.NewLogging(p.sg.ID(), p.sg.Type(), p.sg.Methods(), p.sg.Transport().MethodOptions))
		}
		if p.sg.InstrumentingEnable() {
			generators = append(generators, generator.NewInstrumenting(
				p.sg.ID(),
				p.sg.Type(),
				p.sg.Methods(),
				p.sg.Transport().MethodOptions,
			))
		}
		if p.sg.Transport().JsonRPC.Enable {
			generators = append(generators, generator.NewJsonRPCServer(p.sg.ID(), p.sg.Type(), p.sg.Methods(), p.sg.Transport()))
		} else {
			generators = append(generators, generator.NewRestServer(p.sg.ID(), p.sg.Type(), p.sg.Methods(), p.sg.Transport()))
		}
	}
	if p.sg.Transport().Client.Enable {
		generators = append(generators,
			generator.NewEndpointFactory(p.sg.ID(), p.sg.Methods(), p.sg.Transport()),
			generator.NewClientStruct(p.sg.ID(), p.sg.Methods(), p.sg.Transport()),
		)
		if p.sg.Transport().JsonRPC.Enable {
			generators = append(
				generators,
				generator.NewJsonRPCGoClient(p.sg.ID(), p.sg.Type(), p.sg.Methods(), p.sg.Transport()),
				generator.NewJsonRPCJSClient(p.sg.Methods(), p.sg.Transport(), p.enums, p.sg.Errors()),
			)
		} else {
			generators = append(generators, generator.NewRestGoClient(p.sg.ID(), p.sg.Type(), p.sg.Methods(), p.sg.Transport()))
		}
	}
	if p.sg.Transport().Openapi.Enable {
		generators = append(generators, generator.NewOpenapi(p.sg.Methods(), p.sg.Transport(), p.workDir, p.sg.Errors()))
	}
	return generators
}

func NewService(
	sg gateway.ServiceGateway,
	gi *git.GIT,
	comments *typeutil.Map,
	enums *typeutil.Map,
	workDir string,
	pkg *packages.Package,

) processor.Processor {
	return &serviceProcessor{
		sg:       sg,
		gi:       gi,
		comments: comments,
		enums:    enums,
		workDir:  workDir,
		pkg:      pkg,
	}
}
