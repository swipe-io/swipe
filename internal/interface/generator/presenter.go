package generator

import (
	"context"
	"fmt"
	stdtypes "go/types"
	"strings"

	"github.com/swipe-io/strcase"
	"github.com/swipe-io/swipe/v2/internal/domain/model"
	"github.com/swipe-io/swipe/v2/internal/importer"
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/writer"
)

type presenterOptionsGateway interface {
	Name() string
	Methods() []model.PresenterMethod
}

type presenterGenerator struct {
	writer.GoLangWriter
	options  presenterOptionsGateway
	importer *importer.Importer
}

func (g *presenterGenerator) Prepare(ctx context.Context) error {
	return nil
}

func (g *presenterGenerator) Process(ctx context.Context) error {
	stName := strcase.ToLowerCamel(g.options.Name())

	g.W("type %s struct{}\n", stName)

	for _, method := range g.options.Methods() {
		fromType := stdtypes.TypeString(method.From, g.importer.QualifyPkg)
		toType := stdtypes.TypeString(method.To, g.importer.QualifyPkg)
		g.W("func (p *%s) %s(from %s) %s {\n", stName, method.Method.Name(), fromType, toType)
		g.W("}\n")
	}
	return nil
}

func (g *presenterGenerator) PkgName() string {
	return ""
}

func (g *presenterGenerator) OutputDir() string {
	return ""
}

func (g *presenterGenerator) SetImporter(importer *importer.Importer) {
	fmt.Println(importer)
	g.importer = importer
}

func (g *presenterGenerator) Filename() string {
	prefix := strcase.ToSnake(g.options.Name())
	if !strings.HasSuffix(prefix, "_presenter") {
		prefix = prefix + "_presenter"
	}
	return prefix + "_gen.go"
}

func NewPresenterGenerator(options presenterOptionsGateway) generator.Generator {
	return &presenterGenerator{options: options}
}
