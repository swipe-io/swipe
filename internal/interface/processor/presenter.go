package processor

import (
	ig "github.com/swipe-io/swipe/v2/internal/interface/generator"
	uga "github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
)

type presenterProcessor struct {
	pg  uga.PresenterGateway
	pkg *packages.Package
}

func (g *presenterProcessor) Pkg() *packages.Package {
	return g.pkg
}

func (g *presenterProcessor) Generators() []ug.Generator {
	return []ug.Generator{ig.NewPresenterGenerator(g.pg)}
}

func NewPresenterGatewayProcessor(pg uga.PresenterGateway, pkg *packages.Package) processor.Processor {
	return &presenterProcessor{pg: pg, pkg: pkg}
}
