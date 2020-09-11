package processor

import (
	"github.com/swipe-io/swipe/v2/internal/interface/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/finder"
	uga "github.com/swipe-io/swipe/v2/internal/usecase/gateway"
	ug "github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"github.com/swipe-io/swipe/v2/internal/usecase/processor"
	"golang.org/x/tools/go/packages"
)

type gatewayProcessor struct {
	hg     uga.HTTPGatewayGateway
	finder finder.ServiceFinder
	pkg    *packages.Package
}

func (g *gatewayProcessor) Pkg() *packages.Package {
	return g.pkg
}

func (g *gatewayProcessor) Generators() []ug.Generator {
	return []ug.Generator{
		generator.NewGatewayGenerator(g.hg.Services()),
	}
}

func NewGatewayProcessor(hg uga.HTTPGatewayGateway, pkg *packages.Package) processor.Processor {
	return &gatewayProcessor{hg: hg, pkg: pkg}
}
