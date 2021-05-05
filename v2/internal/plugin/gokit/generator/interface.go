package generator

import (
	"context"

	"github.com/swipe-io/swipe/v2/internal/plugin/gokit/config"

	"github.com/swipe-io/swipe/v2/internal/swipe"

	"github.com/swipe-io/swipe/v2/internal/option"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type InterfaceGenerator struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
}

func (g *InterfaceGenerator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		ifaceTypeName := NameInterface(iface.Named)

		g.w.W("type %s interface {\n", ifaceTypeName)
		for _, m := range ifaceType.Methods {
			g.w.W(m.Name.Origin)
			g.w.W(importer.TypeString(m.Sig))
			g.w.W("\n")
		}
		g.w.W("}\n")
	}
	return g.w.Bytes()
}

func (g *InterfaceGenerator) OutputDir() string {
	return ""
}

func (g *InterfaceGenerator) Filename() string {
	return "interface.go"
}
