package generator

import (
	"bytes"
	"context"
	stdtypes "go/types"

	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type interfaceGenerator struct {
	writer.BaseWriter
	interfaces model.Interfaces
	i          *importer.Importer
}

func (g *interfaceGenerator) Prepare(ctx context.Context) (err error) {
	return nil
}

func (g *interfaceGenerator) Process(ctx context.Context) (err error) {
	for i := 0; i < g.interfaces.Len(); i++ {
		iface := g.interfaces.At(i)

		interfaceType := iface.LcNameWithPrefix() + "Interface"

		g.W("type %s interface {\n", interfaceType)
		for _, m := range iface.Methods() {
			sig := m.T.(*stdtypes.Signature)
			buf := new(bytes.Buffer)
			buf.WriteString(m.Name)
			stdtypes.WriteSignature(buf, sig, g.i.QualifyPkg)
			g.W("%s\n", buf.String())
		}
		g.W("}\n")
	}
	return nil
}

func (g *interfaceGenerator) PkgName() string {
	return ""
}

func (g *interfaceGenerator) OutputDir() string {
	return ""
}

func (g *interfaceGenerator) Filename() string {
	return "interface_gen.go"
}

func (g *interfaceGenerator) SetImporter(i *importer.Importer) {
	g.i = i
}

func NewInterface(interfaces model.Interfaces) generator.Generator {
	return &interfaceGenerator{interfaces: interfaces}
}
