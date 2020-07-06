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
	kitEndpointPkg := g.i.Import("endpoint", "github.com/go-kit/kit/endpoint")
	ioPkg := g.i.Import("io", "io")
	for _, m := range g.o.Methods {
		g.W("func %sEndpointFactory(instance string) (%s.Endpoint, %s.Closer, error) {\n", m.Name, kitEndpointPkg, ioPkg)
		g.W("s, err := NewClient%s%s(instance)\n", g.o.Transport.Prefix, g.o.ID)
		g.WriteCheckErr(func() {
			g.W("return nil, nil, err\n")
		})
		g.W("return make%sEndpoint(s), nil, nil\n", m.Name)
		g.W("\n}\n\n")
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
