package generator

import (
	"context"

	"github.com/swipe-io/swipe/v3/internal/plugin/gokit/config"
	"github.com/swipe-io/swipe/v3/option"
	"github.com/swipe-io/swipe/v3/swipe"
	"github.com/swipe-io/swipe/v3/writer"
)

type InterfaceGenerator struct {
	w          writer.GoWriter
	Interfaces []*config.Interface
	Output     string
	Pkg        string
}

func (g *InterfaceGenerator) Package() string {
	return g.Pkg
}

func (g *InterfaceGenerator) Generate(ctx context.Context) []byte {
	importer := ctx.Value(swipe.ImporterKey).(swipe.Importer)

	g.w.W("type downloader interface {\nContentType() string\nData() []byte\n}\n\n")

	for _, iface := range g.Interfaces {
		ifaceType := iface.Named.Type.(*option.IfaceType)
		ifaceTypeName := NameInterface(iface)
		middlewareChainName := UcNameWithAppPrefix(iface) + "MiddlewareChain"
		middlewareTypeName := IfaceMiddlewareTypeName(iface)

		g.w.W("type %s interface {\n", ifaceTypeName)
		for _, m := range ifaceType.Methods {
			g.w.W(m.Name.Value)
			g.w.W(swipe.TypeString(m.Sig, false, importer))
			g.w.W("\n")
		}
		g.w.W("}\n")

		g.w.W("type %[1]s func(%[2]s) %[2]s\n", middlewareTypeName, ifaceTypeName)

		g.w.W("func %[1]s(outer %[2]s, others ...%[2]s) %[2]s {return func(next %[3]s) %[3]s {\n\t\tfor i := len(others) - 1; i >= 0; i-- {\nnext = others[i](next)\n}\nreturn outer(next)\n}\n}\n", middlewareChainName, middlewareTypeName, ifaceTypeName)
	}
	return g.w.Bytes()
}

func (g *InterfaceGenerator) OutputPath() string {
	return g.Output
}

func (g *InterfaceGenerator) Filename() string {
	return "interface.go"
}
