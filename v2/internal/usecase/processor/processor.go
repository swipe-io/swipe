package processor

import (
	"github.com/swipe-io/swipe/v2/internal/usecase/generator"
	"golang.org/x/tools/go/packages"
)

type Processor interface {
	Generators() []generator.Generator
	Pkg() *packages.Package
}
