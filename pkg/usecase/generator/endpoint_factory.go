package generator

import (
	"context"

	"github.com/swipe-io/swipe/pkg/domain/model"
	"github.com/swipe-io/swipe/pkg/importer"
	"github.com/swipe-io/swipe/pkg/writer"
)

type endpointFactory struct {
	*writer.GoLangWriter
	filename string
	info     model.GenerateInfo
	o        model.ServiceOption
	i        *importer.Importer
}

func (g *endpointFactory) Prepare(ctx context.Context) error {
	return nil
}

func (g *endpointFactory) Process(ctx context.Context) error {
	g.W("type EndpointFactory struct{\n")
	g.W("Options []%sClientOption\n", g.o.ID)
	g.W("Path string\n")
	g.W("}\n\n")

	if len(g.o.Methods) > 0 {
		kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
		ioPkg := g.i.Import("io", "io")
		stringsPkg := g.i.Import("strings", "strings")

		for _, m := range g.o.Methods {
			g.W("func (f *EndpointFactory) %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", m.Name, kitEndpointPkg, ioPkg)
			g.W("if f.Path != \"\"{\n")
			g.W("instance = %[1]s.TrimRight(instance, \"/\") + \"/\" + %[1]s.TrimLeft(f.Path, \"/\")", stringsPkg)
			g.W("}\n")
			g.W("s, err := NewClient%s%s(instance, f.Options...)\n", g.o.Transport.Prefix, g.o.ID)
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
	return g.filename
}

func (g *endpointFactory) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewEndpointFactory(filename string, info model.GenerateInfo, o model.ServiceOption) Generator {
	return &endpointFactory{GoLangWriter: writer.NewGoLangWriter(), filename: filename, info: info, o: o}
}
