package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type endpointFactory struct {
	writer.GoLangWriter
	interfaces model.Interfaces
	prefix     string
	i          *importer.Importer
}

func (g *endpointFactory) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpointFactory) Process(ctx context.Context) error {
	for i := 0; i < g.interfaces.Len(); i++ {
		iface := g.interfaces.At(i)

		epFactoryName := iface.Name() + "ClientEndpointFactory"
		kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
		ioPkg := g.i.Import("io", "io")
		stringsPkg := g.i.Import("strings", "strings")

		g.W("type %s struct{\n", epFactoryName)
		g.W("opts []ClientOption\n")

		g.W("instance string\n")
		g.W("}\n\n")

		for _, m := range iface.Methods() {
			g.W("func (f *%s) %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, m.Name, kitEndpointPkg, ioPkg)
			g.W("if f.instance != \"\"{\n")
			g.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(f.instance, \"/\")", stringsPkg)
			g.W("}\n")
			g.W("c, err :=  NewClient%s(instance, f.opts...)\n", g.prefix)
			g.WriteCheckErr(func() {
				g.W("return nil, nil, err\n")
			})
			g.W("return ")
			g.W("make%sEndpoint(c), nil, nil\n", m.UcName)
			g.W("\n}\n\n")
		}

		g.W("func New%sClientFactory(instance string,", iface.Name())
		g.W("opts ...ClientOption")
		g.W(") *%s {\n", epFactoryName)

		g.W("return &%s{instance: instance, opts: opts}\n", epFactoryName)
		g.W("}\n")
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
	interfaces model.Interfaces,
	prefix string,
) generator.Generator {
	return &endpointFactory{
		interfaces: interfaces,
		prefix:     prefix,
	}
}
