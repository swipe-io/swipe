package generator

import (
	"context"
	stdtypes "go/types"

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

		if iface.External() {
			epFactoryName := iface.LoweName() + "EndpointFactory"
			kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
			ioPkg := g.i.Import("io", "io")
			stringsPkg := g.i.Import("strings", "strings")

			g.W("type %s struct{\n", epFactoryName)
			g.W("factory func(instance string) (%s, error)\n", stdtypes.TypeString(iface.Type(), g.i.QualifyPkg))
			g.W("instance string\n")
			g.W("}\n\n")

			for _, m := range iface.Methods() {
				g.W("func (f *%s) %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", epFactoryName, m.Name, kitEndpointPkg, ioPkg)
				g.W("if f.instance != \"\"{\n")
				g.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(f.instance, \"/\")", stringsPkg)
				g.W("}\n")
				g.W("c, err := f.factory(instance)\n")
				g.WriteCheckErr(func() {
					g.W("return nil, nil, err\n")
				})
				g.W("return ")
				g.W("make%sEndpoint(c), nil, nil\n", m.NameExport)
				g.W("\n}\n\n")
			}

			g.W("func New%sFactory(instance string,", iface.Name())
			g.W("factory func(instance string) (%s, error)", stdtypes.TypeString(iface.Type(), g.i.QualifyPkg))
			g.W(") %sEndpointFactory {\n", iface.Name())

			g.W("return &%s{instance: instance, factory: factory}\n", epFactoryName)
			g.W("}\n")
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
	interfaces model.Interfaces,
	prefix string,
) generator.Generator {
	return &endpointFactory{
		interfaces: interfaces,
		prefix:     prefix,
	}
}
