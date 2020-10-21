package processor

import (
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	uga "github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
)

type configProcessor struct {
	cg      uga.ConfigGateway
	pkg     *packages.Package
	workDir string
}

func (p *configProcessor) Pkg() *packages.Package {
	return p.pkg
}

func (p *configProcessor) Generators() []ug.Generator {
	generators := []ug.Generator{
		generator.NewConfig(p.cg.Struct(), p.cg.StructType(), p.cg.StructExpr(), p.cg.FuncName()),
	}
	if p.cg.DocEnable() {
		generators = append(generators, generator.NewConfigDoc(p.cg.Struct(), p.workDir, p.cg.DocOutputDir()))
	}
	return generators
}

func NewConfig(cg uga.ConfigGateway, pkg *packages.Package, workDir string) processor.Processor {
	return &configProcessor{cg: cg, pkg: pkg, workDir: workDir}
}
