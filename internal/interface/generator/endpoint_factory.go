package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type endpointFactoryOptionsGateway interface {
	Interfaces() model.Interfaces
	Prefix() string
}

type endpointFactory struct {
	writer.GoLangWriter
	options endpointFactoryOptionsGateway
	i       *importer.Importer
}

func (g *endpointFactory) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpointFactory) Process(ctx context.Context) error {

	for i := 0; i < g.options.Interfaces().Len(); i++ {
		iface := g.options.Interfaces().At(i)

		epFactoryName := iface.Name() + "EndpointFactory"

		g.W("type %s struct{\n", epFactoryName)
		g.W("Option []ClientOption\n")
		g.W("Path string\n")
		g.W("}\n\n")

		kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
		ioPkg := g.i.Import("io", "io")
		stringsPkg := g.i.Import("strings", "strings")

		for _, m := range iface.Methods() {
			g.W("func (f *%s) %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, m.Name, kitEndpointPkg, ioPkg)
			g.W("if f.Path != \"\"{\n")
			g.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(f.Path, \"/\")", stringsPkg)
			g.W("}\n")
			g.W("c, err := NewClient%s(instance, f.Option...)\n", g.options.Prefix())
			g.WriteCheckErr(func() {
				g.W("return nil, nil, err\n")
			})
			g.W("return ")
			if g.options.Interfaces().Len() > 1 {
				g.W("make%sEndpoint(c.%sClient), nil, nil\n", m.NameExport, iface.NameExport())
			} else {
				g.W("make%sEndpoint(c), nil, nil\n", m.NameExport)
			}
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
	options endpointFactoryOptionsGateway,
) generator.Generator {
	return &endpointFactory{
		options: options,
	}
}
