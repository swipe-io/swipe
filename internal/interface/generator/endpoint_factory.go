package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/usecase/generator"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type endpointFactory struct {
	writer.GoLangWriter
	i              *importer.Importer
	serviceID      string
	serviceMethods []model.ServiceMethod
	transport      model.TransportOption
}

func (g *endpointFactory) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpointFactory) Process(ctx context.Context) error {
	g.W("type EndpointFactory struct{\n")
	g.W("Option []%sClientOption\n", g.serviceID)
	g.W("Path string\n")
	g.W("}\n\n")

	if len(g.serviceMethods) > 0 {
		kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
		ioPkg := g.i.Import("io", "io")
		stringsPkg := g.i.Import("strings", "strings")

		for _, m := range g.serviceMethods {
			g.W("func (f *EndpointFactory) %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", m.Name, kitEndpointPkg, ioPkg)
			g.W("if f.Path != \"\"{\n")
			g.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(f.Path, \"/\")", stringsPkg)
			g.W("}\n")
			g.W("s, err := NewClient%s%s(instance, f.Option...)\n", g.transport.Prefix, g.serviceID)
			g.WriteCheckErr(func() {
				g.W("return nil, nil, err\n")
			})
			g.W("return make%sEndpoint(s), nil, nil\n", m.Name)
			g.W("\n}\n\n")
		}
	}
	return nil
}

func (g *endpointFactory) PkgName() string {
	return ""
}

func (g *endpointFactory) OutputDir() string {
	return ""
}

func (g *endpointFactory) Filename() string {
	return "endpoint_factory_gen.go"
}

func (g *endpointFactory) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewEndpointFactory(
	serviceID string,
	serviceMethods []model.ServiceMethod,
	transport model.TransportOption,
) generator.Generator {
	return &endpointFactory{
		serviceID:      serviceID,
		serviceMethods: serviceMethods,
		transport:      transport,
	}
}
